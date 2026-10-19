package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "empty.bin", nil)
	pdfxtest.Raw(t, "testdata", "one-byte.bin", []byte{0})
	pdfxtest.Raw(t, "testdata", "sample.txt", pdfxtest.Sample)
	pdfxtest.Raw(t, "testdata", "blank.pdf", pdfxtest.Write(pdfxtest.Pages("blank")))
}
