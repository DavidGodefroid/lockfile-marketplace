package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
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
	for _, r := range []io.Reader{
		iotest.OneByteReader(strings.NewReader("ab\x00c")),
		iotest.HalfReader(bytes.NewReader(late)),
		iotest.DataErrReader(strings.NewReader("abc\x00")),
	} {
		got, err := hasNul(r)
		if err != nil || !got {
			t.Errorf("hasNul = %v, %v; want true, nil", got, err)
		}
	}
}

func TestReadError(t *testing.T) {
	want := errors.New("boom")
	if _, err := hasNul(iotest.ErrReader(want)); !errors.Is(err, want) {
		t.Errorf("hasNul error = %v; want %v", err, want)
	}
}
