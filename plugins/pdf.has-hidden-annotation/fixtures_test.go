package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func annot(extra string) string {
	return "<< /Type /Annot /Subtype /Text /Rect [72 700 144 720] /Contents (sample)" + extra + " >>"
}

func writeFixtures(t *testing.T) {
	content := "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	res := "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"
	page := func(annots ...string) pdfxtest.Page {
		return pdfxtest.Page{Content: content, Resources: res, Annots: annots}
	}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{page()}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	one := func(a ...string) pdfxtest.Doc { return pdfxtest.Doc{Pages: []pdfxtest.Page{page(a...)}} }
	hidden := annot(" /F 2")
	write("plain.pdf", pdfxtest.Doc{})
	write("hidden.pdf", one(hidden))
	write("no-view.pdf", one(annot(" /F 32")))
	write("hidden-and-print.pdf", one(annot(" /F 6")))
	write("hidden-second-page.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(), page(hidden)}})
	write("hidden-object-stream.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(hidden)}, ObjStream: true})
	write("indirect-flags.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot(" /F {{obj 0}}"))}, Objects: []string{"32"}})
	write("after-visible.pdf", one(annot(" /F 4"), hidden))
	write("no-flags.pdf", one(annot("")))
	write("print-only.pdf", one(annot(" /F 4")))
	write("invisible-not-hidden.pdf", one(annot(" /F 1")))
	write("unlisted.pdf", pdfxtest.Doc{Objects: []string{hidden}})
}

// flagsNotInteger is a PDF whose only annotation has a name as /F.
func flagsNotInteger() []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{
		Content:   "BT /F1 12 Tf 72 720 Td (sample) Tj ET",
		Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>",
		Annots:    []string{annot(" /F /Hidden")},
	}}})
}
