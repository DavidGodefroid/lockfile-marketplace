package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const ns = "http://www.aiim.org/pdfua/ns/id/"

const (
	element     = `<rdf:Description rdf:about="" xmlns:pdfuaid="` + ns + `"><pdfuaid:part>1</pdfuaid:part></rdf:Description>`
	attribute   = `<rdf:Description rdf:about="" xmlns:pdfuaid="` + ns + `" pdfuaid:part="1"/>`
	otherPrefix = `<rdf:Description rdf:about="" xmlns:ua="` + ns + `"><ua:part>1</ua:part></rdf:Description>`
	wrongNS     = `<rdf:Description rdf:about="" xmlns:pdfuaid="http://example.org/not-pdfua/"><pdfuaid:part>1</pdfuaid:part></rdf:Description>`
)

func withXMP(x []byte) []byte {
	d := pdfxtest.Pages("text")
	d.Metadata = x
	return pdfxtest.Write(d)
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", pdfxtest.Write(pdfxtest.Pages("text")))
	pdfxtest.Raw(t, "testdata", "xmp-without-pdfuaid.pdf", withXMP(pdfxtest.XMP(pdfxtest.NoPDFA)))
	pdfxtest.Raw(t, "testdata", "pdfa-only.pdf", withXMP(pdfxtest.XMP(pdfxtest.PDFAElement)))
	pdfxtest.Raw(t, "testdata", "pdfuaid-element.pdf", withXMP(pdfxtest.XMP(element)))
	pdfxtest.Raw(t, "testdata", "pdfuaid-attribute.pdf", withXMP(pdfxtest.XMP(attribute)))
	pdfxtest.Raw(t, "testdata", "pdfuaid-other-prefix.pdf", withXMP(pdfxtest.XMP(otherPrefix)))
	pdfxtest.Raw(t, "testdata", "pdfuaid-wrong-namespace.pdf", withXMP(pdfxtest.XMP(wrongNS)))
	pdfxtest.Raw(t, "testdata", "pdfuaid-utf16.pdf", withXMP(pdfxtest.UTF16BE(pdfxtest.XMP(element))))
}
