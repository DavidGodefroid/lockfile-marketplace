package main

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	pdf := pdfxtest.Write(pdfxtest.Pages("blank"))
	pdfxtest.Raw(t, "testdata", "empty.bin", nil)
	pdfxtest.Raw(t, "testdata", "sample.txt", pdfxtest.Sample)
	pdfxtest.Raw(t, "testdata", "blank.pdf", pdf)
	pdfxtest.Raw(t, "testdata", "header-at-1000.bin", append(bytes.Repeat([]byte("x"), 1000), pdf...))
	pdfxtest.Raw(t, "testdata", "header-at-2000.bin", append(bytes.Repeat([]byte("x"), 2000), pdf...))
	pdfxtest.Raw(t, "testdata", "pdf-without-dash.bin", []byte("%PDF1.7\nsample\n"))
}
