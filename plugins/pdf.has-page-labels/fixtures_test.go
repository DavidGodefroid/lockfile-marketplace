package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-labels.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "decimal-labels.pdf", doc(map[string]string{"/PageLabels": "<< /Nums [0 << /S /D >>] >>"}))
	pdfxtest.Raw(t, "testdata", "prefixed-labels.pdf", doc(map[string]string{"/PageLabels": "<< /Nums [0 << /S /r /P (sample-) >>] >>"}))
}
