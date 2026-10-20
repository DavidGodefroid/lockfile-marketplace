package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	pdfProducer     = `<rdf:Description rdf:about="" xmlns:pdf="http://ns.adobe.com/pdf/1.3/"><pdf:Producer>sample</pdf:Producer></rdf:Description>`
	pdfProducerAttr = `<rdf:Description rdf:about="" xmlns:pdf="http://ns.adobe.com/pdf/1.3/" pdf:Producer="sample"/>`
	// otherProducer uses the pdf prefix bound to another namespace: not the Adobe PDF schema.
	otherProducer = `<rdf:Description rdf:about="" xmlns:pdf="http://ns.example.invalid/sample/"><pdf:Producer>sample</pdf:Producer></rdf:Description>`
)

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-producer.pdf", doc(map[string]string{"/Producer": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "blank-producer.pdf", doc(map[string]string{"/Producer": "(   )"}, nil))
	pdfxtest.Raw(t, "testdata", "producer.pdf", doc(map[string]string{"/Producer": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "utf16-hex-producer.pdf", doc(map[string]string{"/Producer": "<FEFF00730061006D0070006C0065>"}, nil))
	pdfxtest.Raw(t, "testdata", "creator-only.pdf", doc(map[string]string{"/Creator": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-pdf-producer.pdf", doc(nil, pdfxtest.XMP(pdfProducer)))
	pdfxtest.Raw(t, "testdata", "xmp-pdf-producer-attr.pdf", doc(nil, pdfxtest.XMP(pdfProducerAttr)))
	pdfxtest.Raw(t, "testdata", "xmp-other-namespace.pdf", doc(nil, pdfxtest.XMP(otherProducer)))
	pdfxtest.Raw(t, "testdata", "xmp-without-producer.pdf", doc(map[string]string{"/Producer": "()"}, pdfxtest.XMP(pdfxtest.NoPDFA)))
}
