package main

import (
	"os"
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

// TestNoAnswer checks that a header declaring more than maxPixels pixels is
// an error, not a verdict.
func TestNoAnswer(t *testing.T) {
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"over pixel limit": sized(8000, 4001),
		"huge":             sized(1<<20, 1<<20),
	})
}
