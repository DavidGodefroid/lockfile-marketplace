package main

import (
	"errors"
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

// TestReadError checks that a read error is returned, not turned into a
// verdict.
func TestReadError(t *testing.T) {
	boom := errors.New("boom")
	if _, err := parses(iotest.ErrReader(boom)); !errors.Is(err, boom) {
		t.Fatalf("parses(ErrReader) error = %v, want %v", err, boom)
	}
}
