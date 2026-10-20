package main

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(linearized bool) []byte {
	d := pdfxtest.Pages("text", "blank")
	d.Linearized = linearized
	return pdfxtest.Write(d)
}

// wrongLength changes the first digit of the /L of a linearized file,
// keeping the file size.
func wrongLength(b []byte) []byte {
	b = bytes.Clone(b)
	i := bytes.Index(b, []byte("/L ")) + 3
	if b[i] == '9' {
		b[i] = '8'
	} else {
		b[i] = '9'
	}
	return b
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "plain.pdf", doc(false))
	pdfxtest.Raw(t, "testdata", "linearized.pdf", doc(true))
	pdfxtest.Raw(t, "testdata", "linearized-then-updated.pdf",
		pdfxtest.AppendUpdate(doc(true), []string{"<< /Producer (sample) >>"}, "/Info {{obj 0}}"))
	pdfxtest.Raw(t, "testdata", "linearized-wrong-length.pdf", wrongLength(doc(true)))
}
