package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	dcDescription = `<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:description><rdf:Alt><rdf:li xml:lang="x-default">sample</rdf:li></rdf:Alt></dc:description></rdf:Description>`
	dcTitle       = `<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title><rdf:Alt><rdf:li xml:lang="x-default">sample</rdf:li></rdf:Alt></dc:title></rdf:Description>`
	otherNS       = `<rdf:Description rdf:about="" xmlns:dc="http://example.com/sample/"><dc:description>sample</dc:description></rdf:Description>`
)

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-subject.pdf", doc(map[string]string{"/Subject": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "blank-subject.pdf", doc(map[string]string{"/Subject": "(   )"}, nil))
	pdfxtest.Raw(t, "testdata", "subject.pdf", doc(map[string]string{"/Subject": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "utf16-hex-subject.pdf", doc(map[string]string{"/Subject": "<FEFF00730061006D0070006C0065>"}, nil))
	pdfxtest.Raw(t, "testdata", "title-only.pdf", doc(map[string]string{"/Title": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-dc-description.pdf", doc(nil, pdfxtest.XMP(dcDescription)))
	pdfxtest.Raw(t, "testdata", "xmp-dc-title-only.pdf", doc(map[string]string{"/Subject": "()"}, pdfxtest.XMP(dcTitle)))
	pdfxtest.Raw(t, "testdata", "xmp-description-other-namespace.pdf", doc(nil, pdfxtest.XMP(otherNS)))
}
