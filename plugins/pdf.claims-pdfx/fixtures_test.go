package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	element = `<rdf:Description rdf:about="" xmlns:pdfxid="` + pdfx.PDFXIDNamespace + `"><pdfxid:GTS_PDFXVersion>PDF/X-4</pdfxid:GTS_PDFXVersion></rdf:Description>`
	wrongNS = `<rdf:Description rdf:about="" xmlns:pdfxid="http://example.org/not-pdfx/"><pdfxid:GTS_PDFXVersion>PDF/X-4</pdfxid:GTS_PDFXVersion></rdf:Description>`
)

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-claim.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "info-key.pdf", doc(map[string]string{"/GTS_PDFXVersion": "(PDF/X-1:2001)"}, nil))
	pdfxtest.Raw(t, "testdata", "info-other-keys.pdf", doc(map[string]string{"/Title": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-property.pdf", doc(nil, pdfxtest.XMP(element)))
	pdfxtest.Raw(t, "testdata", "xmp-wrong-namespace.pdf", doc(nil, pdfxtest.XMP(wrongNS)))
	pdfxtest.Raw(t, "testdata", "xmp-pdfa-only.pdf", doc(nil, pdfxtest.XMP(pdfxtest.PDFAElement)))
}
