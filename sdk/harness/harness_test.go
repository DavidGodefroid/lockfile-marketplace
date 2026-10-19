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

func TestHarnessCatchesHang(t *testing.T) {
	err := Verify("testdata/hang")
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("Verify = %v", err)
	}
}
