package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const filespec = "<< /Type /Filespec /F (sample.xml) /UF (sample.xml) /AFRelationship /Data >>"

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-af.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "empty-af.pdf", doc(map[string]string{"/AF": "[]"}))
	pdfxtest.Raw(t, "testdata", "one-file.pdf", doc(map[string]string{"/AF": "[" + filespec + "]"}))
	pdfxtest.Raw(t, "testdata", "indirect-af.pdf", pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{}},
		Catalog: map[string]string{"/AF": "{{obj 0}}"},
		Objects: []string{"[" + filespec + "]"},
	}))
}
