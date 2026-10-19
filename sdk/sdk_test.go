package sdk

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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
