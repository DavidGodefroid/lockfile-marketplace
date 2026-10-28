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

// A /SubFilter that is not a name is a parse failure, not a 0.
func TestNonNameSubFilter(t *testing.T) {
	ctx, err := pdfx.Open(bytes.NewReader(pdfxtest.Sign(pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT ET"}},
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] /SigFlags 3 >>"},
		Objects: []string{pdfxtest.SignatureField(0), sigValue("/Sig", "(sample)")},
	}))))
	if err != nil {
		t.Fatalf("pdfx.Open: %v", err)
	}
	sigs, err := pdfx.SignatureValues(ctx)
	if err != nil {
		t.Fatalf("pdfx.SignatureValues: %v", err)
	}
	if len(sigs) != 1 {
		t.Fatalf("SignatureValues = %d values; want 1", len(sigs))
	}
	if ts, err := isTimestamp(ctx, sigs[0]); err == nil {
		t.Fatalf("isTimestamp = %v, nil; want an error", ts)
	}
}
