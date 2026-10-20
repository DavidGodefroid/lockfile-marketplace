package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func annot(subtype, extra string) string {
	return "<< /Type /Annot /Subtype /" + subtype + " /Rect [72 700 144 720]" + extra + " >>"
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
	redact := annot("Redact", " /QuadPoints [72 720 144 720 72 700 144 700] /DA (0 g)")
	write("plain.pdf", pdfxtest.Doc{})
	write("redact.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(redact)}})
	write("redact-second-page.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Text", "")), page(annot("Link", ""), redact)}})
	write("object-stream.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(redact)}, ObjStream: true})
	write("other-markup.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Highlight", " /QuadPoints [72 720 144 720 72 700 144 700]"), annot("Square", ""))}})
	write("unlisted.pdf", pdfxtest.Doc{Objects: []string{redact}})
}
