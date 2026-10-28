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
		"zero pages": pdfxtest.Write(pdfxtest.Doc{}),
		"tabs not a name": pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{
			Annots:  []string{"<< /Type /Annot /Subtype /Text /Rect [72 700 144 720] >>"},
			Entries: map[string]string{"/Tabs": "1"},
		}}}),
		"sample.txt": pdfxtest.Sample,
		"empty":      nil,
	})
}
