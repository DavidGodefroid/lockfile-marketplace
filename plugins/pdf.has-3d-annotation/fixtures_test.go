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
	// pdfcpu requires /3DD: a minimal U3D stream, shared as Objects[0].
	model := []string{pdfxtest.Stream("<< /Type /3D /Subtype /U3D >>", pdfxtest.Sample)}
	threeD := annot("3D", " /3DD {{obj 0}}")
	write("plain.pdf", pdfxtest.Doc{})
	write("3d.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(threeD)}, Objects: model})
	write("3d-second-page.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Link", "")), page(annot("Text", ""), threeD)}, Objects: model})
	write("object-stream.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(threeD)}, ObjStream: true, Objects: model})
	write("square-only.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Square", ""))}})
	write("text-note-only.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Text", " /Contents (sample)"))}})
	write("unlisted.pdf", pdfxtest.Doc{Objects: append(model, threeD)})
}
