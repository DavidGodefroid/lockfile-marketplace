package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const tree = "<< /Type /StructTreeRoot >>"

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "untagged.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "tagged.pdf", doc(map[string]string{"/MarkInfo": "<< /Marked true >>", "/StructTreeRoot": tree}))
	pdfxtest.Raw(t, "testdata", "tree-only.pdf", doc(map[string]string{"/StructTreeRoot": tree}))
	pdfxtest.Raw(t, "testdata", "marked-only.pdf", doc(map[string]string{"/MarkInfo": "<< /Marked true >>"}))
	pdfxtest.Raw(t, "testdata", "marked-false.pdf", doc(map[string]string{"/MarkInfo": "<< /Marked false >>", "/StructTreeRoot": tree}))
}
