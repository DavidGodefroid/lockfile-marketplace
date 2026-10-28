package main

import (
	"os"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
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
	bomb := deflate(t, make([]byte, maxDecoded+1))
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt": pdfxtest.Sample,
		"empty":      nil,
		// Decodes past the bound: an error, not an answer.
		"over-bound.pdf": pdf([]string{pdfxtest.Stream("<< /Filter /FlateDecode >>", bomb)}, false),
	})
}
