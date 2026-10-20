package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

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

type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }

func TestReadErrorIsReturned(t *testing.T) {
	want := errors.New("disk failure")
	ok, err := wellFormed(io.MultiReader(strings.NewReader("<a>"), failingReader{want}))
	if !errors.Is(err, want) || ok {
		t.Fatalf("wellFormed = %v, %v; want false, %v", ok, err, want)
	}
}

func TestReadErrorInLatin1IsReturned(t *testing.T) {
	want := errors.New("disk failure")
	r := io.MultiReader(strings.NewReader("<?xml version=\"1.0\" encoding=\"latin1\"?><a>"), failingReader{want})
	ok, err := wellFormed(r)
	if !errors.Is(err, want) || ok {
		t.Fatalf("wellFormed = %v, %v; want false, %v", ok, err, want)
	}
}
