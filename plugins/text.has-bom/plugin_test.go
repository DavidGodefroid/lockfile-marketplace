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
	for _, b := range [][]byte{
		{0xEF, 0xBB, 0xBF, 'a'},
		{0x00, 0x00, 0xFE, 0xFF},
	} {
		got, err := hasBOM(iotest.OneByteReader(bytes.NewReader(b)))
		if err != nil || !got {
			t.Errorf("hasBOM(% X) = %v, %v; want true, nil", b, got, err)
		}
	}
}

func TestReadError(t *testing.T) {
	want := errors.New("boom")
	r := io.MultiReader(bytes.NewReader([]byte{0xEF}), iotest.ErrReader(want))
	if _, err := hasBOM(r); !errors.Is(err, want) {
		t.Errorf("hasBOM error = %v; want %v", err, want)
	}
}
