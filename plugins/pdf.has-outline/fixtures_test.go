package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string, objects ...string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog, Objects: objects})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-outline.pdf", doc(nil))
	// pdfcpu requires /Outlines to be an indirect reference.
	pdfxtest.Raw(t, "testdata", "empty-outline.pdf", doc(
		map[string]string{"/Outlines": "{{obj 0}}"},
		"<< /Type /Outlines /Count 0 >>",
	))
	pdfxtest.Raw(t, "testdata", "one-item.pdf", doc(
		map[string]string{"/Outlines": "{{obj 0}}"},
		"<< /Type /Outlines /First {{obj 1}} /Last {{obj 1}} /Count 1 >>",
		"<< /Title (sample) /Parent {{obj 0}} /Dest [3 0 R /Fit] >>",
	))
}
