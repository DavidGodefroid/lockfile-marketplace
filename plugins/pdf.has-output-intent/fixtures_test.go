package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const intent = "<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier (sample) >>"

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-intent.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "empty-array.pdf", doc(map[string]string{"/OutputIntents": "[]"}))
	pdfxtest.Raw(t, "testdata", "one-intent.pdf", doc(map[string]string{"/OutputIntents": "[" + intent + "]"}))
	pdfxtest.Raw(t, "testdata", "indirect-intent.pdf", pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{}},
		Catalog: map[string]string{"/OutputIntents": "[{{obj 0}}]"},
		Objects: []string{intent},
	}))
}
