package main

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	good := pdfxtest.Write(pdfxtest.Pages("text", "blank"))
	i := bytes.LastIndex(good, []byte("startxref\n"))
	pdfxtest.Raw(t, "testdata", "good.pdf", good)
	pdfxtest.Raw(t, "testdata", "cut-eof.pdf", good[:len(good)-7])
	pdfxtest.Raw(t, "testdata", "cut-half.pdf", good[:len(good)/2])
	pdfxtest.Raw(t, "testdata", "startxref-to-garbage.pdf", append(bytes.Clone(good[:i]), "startxref\n12\n%%EOF\n"...))
	pdfxtest.Raw(t, "testdata", "junk-after-eof.pdf", append(bytes.Clone(good), bytes.Repeat([]byte{0}, 2000)...))
	pdfxtest.Raw(t, "testdata", "eol-after-eof.pdf", append(bytes.Clone(good), "\r\n"...))
}
