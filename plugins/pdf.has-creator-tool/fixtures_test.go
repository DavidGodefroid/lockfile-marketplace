package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	xmpCreatorTool = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/"><xmp:CreatorTool>sample</xmp:CreatorTool></rdf:Description>`
	xmpCreatorAttr = `<rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmp:CreatorTool="sample"/>`
	// pdf:Producer is a different property: it must not count as a creator tool.
	xmpProducerOnly = `<rdf:Description rdf:about="" xmlns:pdf="http://ns.adobe.com/pdf/1.3/"><pdf:Producer>sample</pdf:Producer></rdf:Description>`
)

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-creator.pdf", doc(map[string]string{"/Creator": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "blank-creator.pdf", doc(map[string]string{"/Creator": "(   )"}, nil))
	pdfxtest.Raw(t, "testdata", "creator.pdf", doc(map[string]string{"/Creator": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "utf16-hex-creator.pdf", doc(map[string]string{"/Creator": "<FEFF00730061006D0070006C0065>"}, nil))
	pdfxtest.Raw(t, "testdata", "producer-only.pdf", doc(map[string]string{"/Producer": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-creator-tool.pdf", doc(nil, pdfxtest.XMP(xmpCreatorTool)))
	pdfxtest.Raw(t, "testdata", "xmp-creator-tool-attribute.pdf", doc(nil, pdfxtest.XMP(xmpCreatorAttr)))
	pdfxtest.Raw(t, "testdata", "xmp-producer-only.pdf", doc(map[string]string{"/Creator": "()"}, pdfxtest.XMP(xmpProducerOnly)))
}
