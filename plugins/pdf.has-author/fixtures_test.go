package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const dcCreator = `<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:creator><rdf:Seq><rdf:li>sample</rdf:li></rdf:Seq></dc:creator></rdf:Description>`

func doc(info map[string]string, xmp []byte) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info, Metadata: xmp})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", doc(nil, nil))
	pdfxtest.Raw(t, "testdata", "empty-author.pdf", doc(map[string]string{"/Author": "()"}, nil))
	pdfxtest.Raw(t, "testdata", "blank-author.pdf", doc(map[string]string{"/Author": "(   )"}, nil))
	pdfxtest.Raw(t, "testdata", "author.pdf", doc(map[string]string{"/Author": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "utf16-hex-author.pdf", doc(map[string]string{"/Author": "<FEFF00730061006D0070006C0065>"}, nil))
	pdfxtest.Raw(t, "testdata", "title-only.pdf", doc(map[string]string{"/Title": "(sample)"}, nil))
	pdfxtest.Raw(t, "testdata", "xmp-dc-creator.pdf", doc(nil, pdfxtest.XMP(dcCreator)))
	pdfxtest.Raw(t, "testdata", "xmp-without-creator.pdf", doc(map[string]string{"/Author": "()"}, pdfxtest.XMP(pdfxtest.NoPDFA)))
}
