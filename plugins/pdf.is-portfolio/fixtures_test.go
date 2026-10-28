package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "plain.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "collection.pdf", doc(map[string]string{"/Collection": "<< /Type /Collection /View /D >>"}))
	pdfxtest.Raw(t, "testdata", "empty-collection.pdf", doc(map[string]string{"/Collection": "<< >>"}))
}
