package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	xmpModifyElement   = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/"><xmp:ModifyDate>2000-01-01T00:00:00Z</xmp:ModifyDate></rdf:Description>`
	xmpModifyAttribute = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:ModifyDate="2000-01-01T00:00:00Z"/>`
	xmpCreateOnly      = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/"><xmp:CreateDate>2000-01-01T00:00:00Z</xmp:CreateDate></rdf:Description>`
)

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-mod-date.pdf", doc(map[string]string{"/ModDate": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "mod-date.pdf", doc(map[string]string{"/ModDate": "(D:20000101000000Z)"}, nil))
	pdfxtest.Raw(t, "testdata", "creation-date-only.pdf", doc(map[string]string{"/CreationDate": "(D:20000101000000Z)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-modify-date.pdf", doc(nil, pdfxtest.XMP(xmpModifyElement)))
	pdfxtest.Raw(t, "testdata", "xmp-modify-date-attribute.pdf", doc(nil, pdfxtest.XMP(xmpModifyAttribute)))
	pdfxtest.Raw(t, "testdata", "xmp-create-date-only.pdf", doc(map[string]string{"/ModDate": "()"}, pdfxtest.XMP(xmpCreateOnly)))
}
