package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string, objects ...string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog, Objects: objects})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-dests.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "empty-dests.pdf", doc(
		map[string]string{"/Dests": "{{obj 0}}"},
		"<< >>",
	))
	pdfxtest.Raw(t, "testdata", "catalog-dests.pdf", doc(
		map[string]string{"/Dests": "{{obj 0}}"},
		"<< /sample [3 0 R /Fit] >>",
	))
	pdfxtest.Raw(t, "testdata", "names-without-dests.pdf", doc(
		map[string]string{"/Names": "{{obj 0}}"},
		"<< >>",
	))
	pdfxtest.Raw(t, "testdata", "empty-name-tree.pdf", doc(
		map[string]string{"/Names": "{{obj 0}}"},
		"<< /Dests << /Names [] >> >>",
	))
	pdfxtest.Raw(t, "testdata", "name-tree.pdf", doc(
		map[string]string{"/Names": "{{obj 0}}"},
		"<< /Dests {{obj 1}} >>",
		"<< /Names [(sample) [3 0 R /Fit]] >>",
	))
}
