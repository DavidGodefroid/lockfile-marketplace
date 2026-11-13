package sdk

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

// needle is the typed value of the protocol-2 tests: it must never reach
// stdout, stderr or the detail.
const needle = "secret-needle-42"

// In the child process, LOCKFILE_SDK_TEST selects what the answer function
// does; the parent checks what reached stdout, stderr and the exit status.
func TestMain(m *testing.M) {
	switch os.Getenv("LOCKFILE_SDK_TEST") {
	case "one":
		Main(func(f *os.File) (bool, error) {
			b := make([]byte, 3)
			n, _ := f.Read(b)
			return string(b[:n]) == "abc", nil
		})
	case "zero":
		Main(func(*os.File) (bool, error) { return false, nil })
	case "error":
		Main(func(*os.File) (bool, error) { return true, errors.New("secret detail from the file") })
	case "panic":
		Main(func(*os.File) (bool, error) { panic("secret detail from the file") })
	case "inputs":
		Main(func(*os.File) (bool, error) {
			in, err := Inputs()
			return in["needle"] == needle, err
		})
	case "inputs-error":
		Main(func(*os.File) (bool, error) {
			in, _ := Inputs()
			return true, errors.New("bad value " + in["needle"])
		}, WithDetail())
	case "inputs-panic":
		Main(func(*os.File) (bool, error) {
			in, _ := Inputs()
			panic("bad value " + in["needle"])
		}, WithDetail())
	case "detail":
		Main(func(*os.File) (bool, error) {
			Detail().Write(bytes.Repeat([]byte("d"), MaxDetail+1000))
			return true, nil
		}, WithDetail())
	case "detail-undeclared":
		Main(func(*os.File) (bool, error) {
			Detail().Write([]byte("leaked"))
			return true, nil
		})
	}
	os.Exit(m.Run())
}

func child(t *testing.T, mode string) (stdout, stderr string, code int) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "input")
	os.WriteFile(p, []byte("abc"), 0o600)
	in, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "LOCKFILE_SDK_TEST="+mode)
	cmd.ExtraFiles = []*os.File{in}
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err = cmd.Run()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return out.String(), errb.String(), cmd.ProcessState.ExitCode()
}

func TestMainWritesOne(t *testing.T) {
	out, errs, code := child(t, "one")
	if out != "1" || errs != "" || code != 0 {
		t.Fatalf("stdout %q stderr %q exit %d", out, errs, code)
	}
}

func TestMainWritesZero(t *testing.T) {
	out, _, code := child(t, "zero")
	if out != "0" || code != 0 {
		t.Fatalf("stdout %q exit %d", out, code)
	}
}

func TestMainErrorWritesNothing(t *testing.T) {
	for _, mode := range []string{"error", "panic"} {
		out, errs, code := child(t, mode)
		if out != "" || errs != "" || code != 1 {
			t.Fatalf("%s: stdout %q stderr %q exit %d", mode, out, errs, code)
		}
	}
}

// child2 runs the child as the sandbox runs a protocol-2 plugin: inputs
// (when not nil) on fd 6 as a sealed read-only memfd, and a pipe on fd 7
// whose content it returns as detail.
func child2(t *testing.T, mode, protocol string, inputs []byte) (stdout, stderr, detail string, code int) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "input")
	os.WriteFile(p, []byte("abc"), 0o600)
	in, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	extra := []*os.File{in, nil, nil, nil}
	if inputs != nil {
		extra[3] = sealedMemfd(t, inputs)
		defer extra[3].Close()
	}
	detR, detW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer detR.Close()
	extra = append(extra, detW)
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = []string{"LOCKFILE_SDK_TEST=" + mode, "LOCKFILE_PLUGIN_PROTOCOL=" + protocol}
	cmd.ExtraFiles = extra
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	detW.Close()
	det, _ := io.ReadAll(detR)
	err = cmd.Wait()
	var ee *exec.ExitError
	if err != nil && !errors.As(err, &ee) {
		t.Fatal(err)
	}
	return out.String(), errb.String(), string(det), cmd.ProcessState.ExitCode()
}

// sealedMemfd is what the sandbox puts on fd 6: a memfd holding b, sealed
// against any change, opened read-only.
func sealedMemfd(t *testing.T, b []byte) *os.File {
	t.Helper()
	fd, err := unix.MemfdCreate("inputs", unix.MFD_CLOEXEC|unix.MFD_ALLOW_SEALING)
	if err != nil {
		t.Skipf("memfd: %v", err)
	}
	f := os.NewFile(uintptr(fd), "inputs")
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}
	seals := unix.F_SEAL_WRITE | unix.F_SEAL_GROW | unix.F_SEAL_SHRINK | unix.F_SEAL_SEAL
	if _, err := unix.FcntlInt(f.Fd(), unix.F_ADD_SEALS, seals); err != nil {
		t.Fatal(err)
	}
	ro, err := os.Open("/proc/self/fd/" + strconv.Itoa(fd))
	if err != nil {
		t.Fatal(err)
	}
	return ro
}

func TestInputsReadsFD6(t *testing.T) {
	out, errs, _, code := child2(t, "inputs", "2", []byte(`{"needle":"`+needle+`"}`))
	if out != "1" || errs != "" || code != 0 {
		t.Fatalf("stdout %q stderr %q exit %d", out, errs, code)
	}
}

func TestInputsFailsSilently(t *testing.T) {
	for name, c := range map[string]struct {
		protocol string
		inputs   []byte
	}{
		"protocol 1":     {"1", []byte(`{"needle":"` + needle + `"}`)},
		"no protocol":    {"", []byte(`{"needle":"` + needle + `"}`)},
		"no fd 6":        {"2", nil},
		"malformed JSON": {"2", []byte(`{"needle":"` + needle)},
		"not strings":    {"2", []byte(`{"needle":42}`)},
		"not an object":  {"2", []byte(`"` + needle + `"`)},
		"null":           {"2", []byte(`null`)},
	} {
		out, errs, _, code := child2(t, "inputs", c.protocol, c.inputs)
		if out != "" || errs != "" || code != 1 {
			t.Errorf("%s: stdout %q stderr %q exit %d", name, out, errs, code)
		}
	}
}

func TestInputsValueNeverPrinted(t *testing.T) {
	for _, mode := range []string{"inputs-error", "inputs-panic"} {
		out, errs, det, code := child2(t, mode, "2", []byte(`{"needle":"`+needle+`"}`))
		if out != "" || errs != "" || det != "" || code != 1 {
			t.Errorf("%s: stdout %q stderr %q detail %q exit %d", mode, out, errs, det, code)
		}
	}
}

func TestDetailReachesFD7Capped(t *testing.T) {
	out, errs, det, code := child2(t, "detail", "2", nil)
	if out != "1" || errs != "" || code != 0 {
		t.Fatalf("stdout %q stderr %q exit %d", out, errs, code)
	}
	if det != strings.Repeat("d", MaxDetail) {
		t.Fatalf("detail: %d bytes, want %d", len(det), MaxDetail)
	}
}

func TestDetailUntouchedWhenUndeclared(t *testing.T) {
	out, _, det, code := child2(t, "detail-undeclared", "2", nil)
	if out != "1" || code != 0 || det != "" {
		t.Fatalf("stdout %q exit %d, fd 7 got %q", out, code, det)
	}
}

func TestDetailNeedsProtocol2(t *testing.T) {
	out, _, det, code := child2(t, "detail", "1", nil)
	if out != "1" || code != 0 || det != "" {
		t.Fatalf("stdout %q exit %d, fd 7 got %q", out, code, det)
	}
}
