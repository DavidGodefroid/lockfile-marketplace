package main

import (
	"errors"
	"os"
	"testing"
	"testing/iotest"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk/harness"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestConformance(t *testing.T) {
	writeFixtures(t)
	harness.Check(t, ".")
}

func TestNoAnswer(t *testing.T) {
	deep := arrays(65, "1")
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":            nil,
		"whitespace":       []byte(" \n\t"),
		"sample.txt":       fx.Sample,
		"sample.docx":      fx.Docx(fx.Paragraph),
		"bom":              []byte("\xef\xbb\xbf[1]"),
		"invalid utf8":     []byte("[\"\xff\"]"),
		"two values":       []byte("[1] [2]"),
		"trailing garbage": []byte("[1] x"),
		"trailing comma":   []byte("[1,]"),
		"deep truncated":   deep[:len(deep)-1],
		"deep then broken": append(deep, " x"...),
	})
}

// TestReadError checks that a read error is returned, not turned into a
// verdict.
func TestReadError(t *testing.T) {
	boom := errors.New("boom")
	if _, err := nestingOver(iotest.ErrReader(boom), maxDepth); !errors.Is(err, boom) {
		t.Fatalf("nestingOver(ErrReader) error = %v, want %v", err, boom)
	}
}
