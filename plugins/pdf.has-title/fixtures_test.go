package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const dcTitle = `<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title><rdf:Alt><rdf:li xml:lang="x-default">sample</rdf:li></rdf:Alt></dc:title></rdf:Description>`

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-title.pdf", doc(map[string]string{"/Title": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "blank-title.pdf", doc(map[string]string{"/Title": "(   )"}, nil))
	pdfxtest.Raw(t, "testdata", "title.pdf", doc(map[string]string{"/Title": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "utf16-hex-title.pdf", doc(map[string]string{"/Title": "<FEFF00730061006D0070006C0065>"}, nil))
	pdfxtest.Raw(t, "testdata", "author-only.pdf", doc(map[string]string{"/Author": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-dc-title.pdf", doc(nil, pdfxtest.XMP(dcTitle)))
	pdfxtest.Raw(t, "testdata", "xmp-without-title.pdf", doc(map[string]string{"/Title": "()"}, pdfxtest.XMP(pdfxtest.NoPDFA)))
}
