package main

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	good := pdfxtest.Write(pdfxtest.Pages("text"))
	after := func(tail []byte) []byte { return append(bytes.Clone(good), tail...) }
	pdfxtest.Raw(t, "testdata", "clean.pdf", good)
	pdfxtest.Raw(t, "testdata", "no-final-eol.pdf", bytes.TrimRight(good, "\n"))
	pdfxtest.Raw(t, "testdata", "eol-after-eof.pdf", after([]byte("\r\n")))
	pdfxtest.Raw(t, "testdata", "nul-padding.pdf", after(make([]byte, 512)))
	pdfxtest.Raw(t, "testdata", "incremental-update.pdf", pdfxtest.AppendUpdate(good, []string{"<< /Sample true >>"}, ""))
	pdfxtest.Raw(t, "testdata", "text-after-eof.pdf", after([]byte("sample\n")))
	pdfxtest.Raw(t, "testdata", "binary-after-eof.pdf", after([]byte{0x01, 0x02, 0x03}))
	// Junk followed by a long whitespace run: the junk is more than one
	// scan chunk away from the end of the file.
	pdfxtest.Raw(t, "testdata", "junk-before-long-padding.pdf", after(append([]byte("sample"), bytes.Repeat([]byte(" "), 3*chunk)...)))
}
