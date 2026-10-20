package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	xmpCreateElement   = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/"><xmp:CreateDate>2000-01-01T00:00:00Z</xmp:CreateDate></rdf:Description>`
	xmpCreateAttribute = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:CreateDate="2000-01-01T00:00:00Z"/>`
	xmpModifyOnly      = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/"><xmp:ModifyDate>2000-01-01T00:00:00Z</xmp:ModifyDate></rdf:Description>`
)

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-creation-date.pdf", doc(map[string]string{"/CreationDate": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "creation-date.pdf", doc(map[string]string{"/CreationDate": "(D:20000101000000Z)"}, nil))
	pdfxtest.Raw(t, "testdata", "mod-date-only.pdf", doc(map[string]string{"/ModDate": "(D:20000101000000Z)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-create-date.pdf", doc(nil, pdfxtest.XMP(xmpCreateElement)))
	pdfxtest.Raw(t, "testdata", "xmp-create-date-attribute.pdf", doc(nil, pdfxtest.XMP(xmpCreateAttribute)))
	pdfxtest.Raw(t, "testdata", "xmp-modify-date-only.pdf", doc(map[string]string{"/CreationDate": "()"}, pdfxtest.XMP(xmpModifyOnly)))
}
