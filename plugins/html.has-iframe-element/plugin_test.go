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
		iotest.OneByteReader(strings.NewReader("ab<iFrAmE>")),
		iotest.HalfReader(bytes.NewReader(split)),
	} {
		got, err := hasIframe(r)
		if err != nil || !got {
			t.Errorf("hasIframe = %v, %v; want true, nil", got, err)
		}
	}
}

func TestReadError(t *testing.T) {
	want := errors.New("boom")
	if _, err := hasIframe(iotest.ErrReader(want)); !errors.Is(err, want) {
		t.Errorf("hasIframe error = %v; want %v", err, want)
	}
}
