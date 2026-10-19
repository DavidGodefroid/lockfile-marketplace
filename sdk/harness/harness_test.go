package harness

import (
	"os"
	"strings"
	"testing"

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
