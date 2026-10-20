package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(lang string) []byte {
	d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}
	if lang != "" {
		d.Catalog = map[string]string{"/Lang": lang}
	}
	return pdfxtest.Write(d)
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-lang.pdf", doc(""))
	pdfxtest.Raw(t, "testdata", "empty-lang.pdf", doc("()"))
	pdfxtest.Raw(t, "testdata", "lang.pdf", doc("(en)"))
	pdfxtest.Raw(t, "testdata", "hex-lang.pdf", doc("<656E>"))
}
