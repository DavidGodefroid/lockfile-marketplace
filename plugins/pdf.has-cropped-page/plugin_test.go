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
		"crop box of three numbers": pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{Entries: map[string]string{"/CropBox": "[0 0 100]"}}}}),
		"crop box not numbers":      pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{Entries: map[string]string{"/CropBox": "[0 0 /A /B]"}}}}),
		"crop box outside media":    pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: box(700, 900, 800, 1000)}}}),
		"no media box":              pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: &pdfxtest.A5}}, NoMediaBox: true}),
		"zero pages":                pdfxtest.Write(pdfxtest.Doc{}),
		"sample.txt":                pdfxtest.Sample,
	})
}
