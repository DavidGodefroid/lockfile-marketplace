package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
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
	})
}

// A /Ff that is not an integer is a parse failure, not a 0: pdfx.Open's
// validation rejects it, and hasReadOnlyField would too.
func TestNonIntegerFlags(t *testing.T) {
	ctx, err := pdfx.Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT ET"}},
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] >>"},
		Objects: []string{"<< /FT /Tx /T (sample) /Ff (sample) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
	})))
	if err != nil {
		return
	}
	if found, err := hasReadOnlyField(ctx); err == nil {
		t.Fatalf("hasReadOnlyField = %v, nil; want an error", found)
	}
}
