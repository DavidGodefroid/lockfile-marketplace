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
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt":      pdfxtest.Sample,
		"empty":           nil,
		"missing-obj.pdf": pdfxtest.AppendUpdate(plain, nil, "/Encrypt 999 0 R"),
		"not-a-dict.pdf":  pdfxtest.AppendUpdate(plain, nil, "/Encrypt 5"),
		"not-boolean.pdf": pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+encrypt("/EncryptMetadata 0")),
		"name-false.pdf":  pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+encrypt("/EncryptMetadata /false")),
	})
}
