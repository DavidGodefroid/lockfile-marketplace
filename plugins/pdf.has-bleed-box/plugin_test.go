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
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"bleed box not an array": pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{Entries: bleed("/Sample")}}}),
		"zero pages":             pdfxtest.Write(pdfxtest.Doc{}),
		"sample.txt":             pdfxtest.Sample,
	})
}
