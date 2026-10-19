package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func withXMP(x []byte) []byte {
	d := pdfxtest.Pages("text")
	d.Metadata = x
	return pdfxtest.Write(d)
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", pdfxtest.Write(pdfxtest.Pages("text")))
	pdfxtest.Raw(t, "testdata", "xmp-without-pdfaid.pdf", withXMP(pdfxtest.XMP(pdfxtest.NoPDFA)))
	pdfxtest.Raw(t, "testdata", "pdfaid-element.pdf", withXMP(pdfxtest.XMP(pdfxtest.PDFAElement)))
	pdfxtest.Raw(t, "testdata", "pdfaid-attribute.pdf", withXMP(pdfxtest.XMP(pdfxtest.PDFAAttribute)))
	pdfxtest.Raw(t, "testdata", "pdfaid-other-prefix.pdf", withXMP(pdfxtest.XMP(pdfxtest.PDFAOtherPrefix)))
	pdfxtest.Raw(t, "testdata", "pdfaid-wrong-namespace.pdf", withXMP(pdfxtest.XMP(pdfxtest.PDFAWrongNS)))
	pdfxtest.Raw(t, "testdata", "pdfaid-utf16.pdf", withXMP(pdfxtest.UTF16BE(pdfxtest.XMP(pdfxtest.PDFAElement))))
}
