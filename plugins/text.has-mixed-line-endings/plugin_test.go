package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"testing/iotest"

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

func TestSplitReads(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"a\r\nb\r\n", false},
		{"a\r\nb\n", true},
		{"a\r\nb\r", true},
		{"a\r\r\n", true},
	} {
		got, err := hasMixedLineEndings(iotest.OneByteReader(bytes.NewReader([]byte(tc.in))))
		if err != nil || got != tc.want {
			t.Errorf("hasMixedLineEndings(%q) = %v, %v; want %v, nil", tc.in, got, err, tc.want)
		}
	}
}

func TestReadError(t *testing.T) {
	want := errors.New("boom")
	r := io.MultiReader(bytes.NewReader([]byte("a\n")), iotest.ErrReader(want))
	if _, err := hasMixedLineEndings(r); !errors.Is(err, want) {
		t.Errorf("hasMixedLineEndings error = %v; want %v", err, want)
	}
}
