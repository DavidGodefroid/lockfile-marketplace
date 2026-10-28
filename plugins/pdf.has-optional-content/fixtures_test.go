package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string, objects ...string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog, Objects: objects})
}

// ocgsNotArray has an /OCGs entry that is a dictionary, not an array.
var ocgsNotArray = doc(
	map[string]string{"/OCProperties": "<< /OCGs {{obj 0}} /D << >> >>"},
	"<< /Type /OCG /Name (sample) >>",
)

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-oc.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "empty-ocgs.pdf", doc(
		map[string]string{"/OCProperties": "<< /OCGs [] /D << >> >>"},
	))
	pdfxtest.Raw(t, "testdata", "one-ocg.pdf", doc(
		map[string]string{"/OCProperties": "<< /OCGs [{{obj 0}}] /D << /ON [{{obj 0}}] >> >>"},
		"<< /Type /OCG /Name (sample) >>",
	))
}
