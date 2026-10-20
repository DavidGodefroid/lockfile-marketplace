package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(header, catalog string) []byte {
	d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Version: header}
	if catalog != "" {
		d.Catalog = map[string]string{"/Version": catalog}
	}
	return pdfxtest.Write(d)
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "header-1.4.pdf", doc("1.4", ""))
	pdfxtest.Raw(t, "testdata", "header-1.7.pdf", doc("1.7", ""))
	pdfxtest.Raw(t, "testdata", "header-2.0.pdf", doc("2.0", ""))
	pdfxtest.Raw(t, "testdata", "header-1.4-catalog-1.7.pdf", doc("1.4", "/1.7"))
	pdfxtest.Raw(t, "testdata", "header-1.7-catalog-2.0.pdf", doc("1.7", "/2.0"))
}
