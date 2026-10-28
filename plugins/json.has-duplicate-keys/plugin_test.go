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

// TestNoAnswer checks that input which is not exactly one JSON value is an
// error, even when a duplicate key was already seen.
func TestNoAnswer(t *testing.T) {
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":              nil,
		"whitespace":         []byte(" \n\t"),
		"sample.txt":         fx.Sample,
		"sample.docx":        fx.Docx(fx.Paragraph),
		"bom":                []byte("\xef\xbb\xbf{\"a\": 1}"),
		"invalid utf8":       []byte("{\"a\": \"\xff\"}"),
		"two values":         []byte("{\"a\": 1}\n{\"a\": 2}"),
		"dup then 2nd value": []byte("{\"a\": 1, \"a\": 2} {}"),
		"dup then truncated": []byte("{\"a\": 1, \"a\": 2"),
		"dup then garbage":   []byte("{\"a\": 1, \"a\": 2} x"),
		"missing colon":      []byte("{\"a\" 1}"),
		"trailing comma":     []byte("{\"a\": 1,}"),
		"unquoted key":       []byte("{a: 1}"),
		"unbalanced close":   []byte("[1]]"),
	})
}

// TestReadError checks that a read error is returned, not turned into a
// verdict.
func TestReadError(t *testing.T) {
	boom := errors.New("boom")
	if _, err := hasDuplicateKeys(iotest.ErrReader(boom)); !errors.Is(err, boom) {
		t.Fatalf("hasDuplicateKeys(ErrReader) error = %v, want %v", err, boom)
	}
}
