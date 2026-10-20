package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	content := "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	res := "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"
	note := "<< /Type /Annot /Subtype /Text /Rect [72 700 144 720] /Contents (sample) >>"
	page := func(tabs string, annots ...string) pdfxtest.Page {
		p := pdfxtest.Page{Content: content, Resources: res, Annots: annots}
		if tabs != "" {
			p.Entries = map[string]string{"/Tabs": tabs}
		}
		return p
	}
	write := func(name string, d pdfxtest.Doc) {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	pages := func(p ...pdfxtest.Page) pdfxtest.Doc { return pdfxtest.Doc{Pages: p} }
	write("plain.pdf", pages(page("")))
	write("tabs-s.pdf", pages(page("/S", note)))
	write("no-tabs.pdf", pages(page("", note)))
	write("tabs-row.pdf", pages(page("/R", note)))
	write("tabs-column.pdf", pages(page("/C", note)))
	write("second-page-lacks.pdf", pages(page("/S", note), page("", note)))
	write("bare-page-without-annots.pdf", pages(page("/S", note), page("")))
	write("empty-annots.pdf", pages(pdfxtest.Page{Content: content, Resources: res, Entries: map[string]string{"/Annots": "[]"}}))
	write("tabs-indirect.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page("{{obj 0}}", note)}, Objects: []string{"/S"}})
	write("object-stream.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page("/S", note)}, ObjStream: true})
	write("unlisted.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page("")}, Objects: []string{note}})
}
