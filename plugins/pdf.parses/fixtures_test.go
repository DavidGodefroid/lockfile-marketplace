package main

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "blank.pdf", pdfxtest.Write(pdfxtest.Pages("blank")))
	pdfxtest.Raw(t, "testdata", "text.pdf", pdfxtest.Write(pdfxtest.Pages("text", "image")))
	pdfxtest.Raw(t, "testdata", "empty.bin", nil)
	pdfxtest.Raw(t, "testdata", "sample.txt", pdfxtest.Sample)
	pdfxtest.Raw(t, "testdata", "garbage.pdf", append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{0xff, 0x00, 'x'}, 300)...))
}
