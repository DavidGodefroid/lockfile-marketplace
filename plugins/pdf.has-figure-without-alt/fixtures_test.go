package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// tagged returns a tagged PDF whose structure tree root is {{obj 0}}, with
// root entries rootEntries, followed by the objects elems ({{obj 1}}…).
func tagged(rootEntries string, objStream bool, elems ...string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{
		Pages: []pdfxtest.Page{{}},
		Catalog: map[string]string{
			"/MarkInfo":       "<< /Marked true >>",
			"/StructTreeRoot": "{{obj 0}}",
		},
		Objects:   append([]string{"<< /Type /StructTreeRoot" + rootEntries + " >>"}, elems...),
		ObjStream: objStream,
	})
}

// nested puts figure under a Document element, next to a paragraph.
func nested(rootEntries string, objStream bool, figure string) []byte {
	return tagged(" /K {{obj 1}}"+rootEntries, objStream,
		"<< /Type /StructElem /S /Document /P {{obj 0}} /K [{{obj 2}} {{obj 3}}] >>",
		"<< /Type /StructElem /S /P /P {{obj 1}} /K 0 >>",
		"<< /Type /StructElem /P {{obj 1}} "+figure+" >>",
	)
}

func writeFixtures(t *testing.T) {
	w := func(name string, b []byte) { pdfxtest.Raw(t, "testdata", name, b) }
	w("untagged.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}))
	w("empty-tree.pdf", tagged("", false))
	w("no-figure.pdf", nested("", false, "/S /Table"))
	w("figure-without-alt.pdf", nested("", false, "/S /Figure"))
	w("figure-with-alt.pdf", nested("", false, "/S /Figure /Alt (sample)"))
	w("figure-with-empty-alt.pdf", nested("", false, "/S /Figure /Alt ()"))
	w("figure-with-actual-text.pdf", nested("", false, "/S /Figure /ActualText (sample)"))
	w("figure-null-alt.pdf", nested("", false, "/S /Figure /Alt null"))
	w("mapped-figure-without-alt.pdf", nested(" /RoleMap << /Pic /Image /Image /Figure >>", false, "/S /Pic"))
	w("mapped-figure-with-alt.pdf", nested(" /RoleMap << /Pic /Figure >>", false, "/S /Pic /Alt (sample)"))
	w("figure-mapped-away.pdf", nested(" /RoleMap << /Figure /Div >>", false, "/S /Figure"))
	w("object-stream.pdf", nested("", true, "/S /Figure"))
	w("second-figure-without-alt.pdf", tagged(" /K [{{obj 1}} {{obj 2}}]", false,
		"<< /S /Figure /P {{obj 0}} /Alt (sample) >>",
		"<< /S /Figure /P {{obj 0}} >>",
	))
}
