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
		"sample.txt": pdfxtest.Sample,
		"empty":      nil,
		// /Ff is a name, not an integer.
		"bad-flags.pdf": pdfxtest.Write(pdfxtest.Doc{
			Pages:   []pdfxtest.Page{{Content: "BT ET"}},
			Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] >>"},
			Objects: []string{"<< /FT /Tx /T (sample) /Ff /Required /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
		}),
	})
}
