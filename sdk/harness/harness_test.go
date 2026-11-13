package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile/pluginspec"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestHarnessGoodPlugin(t *testing.T) {
	Check(t, "testdata/good")
}

func TestHarnessNeedsBothAnswers(t *testing.T) {
	err := Verify("testdata/onlyone")
	if err == nil || !strings.Contains(err.Error(), "answer 0") {
		t.Fatalf("Verify = %v", err)
	}
}

func TestHarnessCatchesWrongAnswer(t *testing.T) {
	err := Verify("testdata/wrong")
	if err == nil || !strings.Contains(err.Error(), "no.txt") {
		t.Fatalf("Verify = %v", err)
	}
}

func TestNoAnswer(t *testing.T) {
	if err := NoAnswer("testdata/fails", map[string][]byte{"empty": nil, "text": []byte("No")}); err != nil {
		t.Fatal(err)
	}
	err := NoAnswer("testdata/fails", map[string][]byte{"answered": []byte("Yes")})
	if err == nil || !strings.Contains(err.Error(), "answered") {
		t.Fatalf("NoAnswer on an answered input = %v", err)
	}
	err = NoAnswer("testdata/hang", map[string][]byte{"hangs": []byte("other")})
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("NoAnswer on a hang = %v", err)
	}
}

func TestHarnessCatchesHang(t *testing.T) {
	err := Verify("testdata/hang")
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("Verify = %v", err)
	}
}

func TestHarnessProtocol2Plugin(t *testing.T) {
	Check(t, "testdata/contains")
}

func TestNoAnswerProtocol2(t *testing.T) {
	// test.contains answers on any file; with its first case's values it
	// answers here too, which NoAnswer reports.
	err := NoAnswer("testdata/contains", map[string][]byte{"answered": []byte("zebra-77")})
	if err == nil || !strings.Contains(err.Error(), "answered true") {
		t.Fatalf("NoAnswer = %v", err)
	}
}

func TestHarnessRefusesMissingInput(t *testing.T) {
	err := Verify("testdata/missinginput")
	if err == nil || !strings.Contains(err.Error(), "no.txt: no value for input needle") {
		t.Fatalf("Verify = %v", err)
	}
}

func TestHarnessRefusesUnfitOrUnknownInput(t *testing.T) {
	m, err := readManifest(mustAbs(t, "testdata/contains"))
	if err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		vals map[string]string
		want string
	}{
		"too long": {map[string]string{"needle": strings.Repeat("x", 65)}, "does not fit"},
		"control":  {map[string]string{"needle": "line\nbreak"}, "does not fit"},
		"unknown":  {map[string]string{"needle": "fine", "other": "extra"}, "other is not an input"},
	} {
		errs := checkValues("expected.toml", Case{File: "f", Inputs: c.vals}, m)
		if len(errs) != 1 || !strings.Contains(errs[0].Error(), c.want) {
			t.Errorf("%s: %v", name, errs)
			continue
		}
		for _, v := range c.vals {
			if strings.Contains(errs[0].Error(), v) {
				t.Errorf("%s: the error quotes a value: %v", name, errs[0])
			}
		}
	}
}

func TestHarnessRefusesEchoedValue(t *testing.T) {
	err := Verify("testdata/echo")
	if err == nil || !strings.Contains(err.Error(), "echoes the value of input needle") {
		t.Fatalf("Verify = %v", err)
	}
}

// The value reaches the plugin on fd 6 only (test.contains fails when it
// finds it in argv or env), and the detail is captured, capped at 4 KiB.
func TestRunPassesInputsAndCapsDetail(t *testing.T) {
	dir := mustAbs(t, "testdata/contains")
	m, err := readManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := build(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	vals := map[string]string{"needle": "zebra-77"}
	var detail bytes.Buffer
	o, _, err := run(context.Background(), bin, bytes.Repeat([]byte("zebra-77 "), 2000), m, vals, &detail)
	if err != nil || o.Answer == nil || !*o.Answer {
		t.Fatalf("answer %v, cause %s, err %v", o.Answer, o.Cause, err)
	}
	if detail.Len() != pluginspec.MaxDetail || !strings.HasPrefix(detail.String(), "found at offset 0\n") {
		t.Fatalf("detail: %d bytes, starts %.40q", detail.Len(), detail.String())
	}
	if strings.Contains(detail.String(), "zebra") {
		t.Fatal("the detail echoes the value")
	}
	detail.Reset()
	o, _, err = run(context.Background(), bin, []byte("no match"), m, vals, &detail)
	if err != nil || o.Answer == nil || *o.Answer || detail.Len() != 0 {
		t.Fatalf("answer %v, cause %s, err %v, detail %q", o.Answer, o.Cause, err, detail.String())
	}
}

func TestRandomValuesFit(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, kind := range []string{pluginspec.KindText, pluginspec.KindEmail, pluginspec.KindDigits} {
		for _, max := range []int{1, 2, 5, 6, 64, pluginspec.MaxInputBytes} {
			in := pluginspec.Input{Name: "v", Kind: kind, MaxBytes: max}
			for range 200 {
				v, ok := randomValue(rng, in)
				if !ok {
					if kind == pluginspec.KindEmail && max < 5 {
						continue
					}
					t.Fatalf("%s/%d: no value", kind, max)
				}
				if in.CheckValue([]byte(v)) != nil {
					t.Fatalf("%s/%d: %q does not fit", kind, max, v)
				}
			}
		}
	}
}

func TestCanonicalJSON(t *testing.T) {
	vals := map[string]string{"b": `quote " back \ é`, "a": "plain"}
	b := canonical(vals)
	if !bytes.HasPrefix(b, []byte(`{"a":"plain","b":`)) {
		t.Fatalf("not sorted: %s", b)
	}
	var got map[string]string
	if err := json.Unmarshal(b, &got); err != nil || got["a"] != vals["a"] || got["b"] != vals["b"] {
		t.Fatalf("round trip: %v %v", got, err)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
