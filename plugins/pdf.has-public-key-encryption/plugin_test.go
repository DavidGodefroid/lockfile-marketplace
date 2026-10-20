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

// TestNoAnswer: a file that is not a PDF, or whose /Encrypt or its
// /Filter cannot be read, gets no answer.
func TestNoAnswer(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt":         pdfxtest.Sample,
		"empty":              nil,
		"encrypt-int.pdf":    pdfxtest.AppendUpdate(plain, nil, "/Encrypt 5"),
		"encrypt-dangle.pdf": pdfxtest.AppendUpdate(plain, nil, "/Encrypt 999 0 R"),
		"no-filter.pdf":      pdfxtest.AppendUpdate(plain, []string{"<< /V 4 /R 4 >>"}, "/Encrypt {{obj 0}}"),
		"filter-string.pdf":  pdfxtest.AppendUpdate(plain, []string{"<< /Filter (Adobe.PubSec) /V 4 >>"}, "/Encrypt {{obj 0}}"),
	})
}
