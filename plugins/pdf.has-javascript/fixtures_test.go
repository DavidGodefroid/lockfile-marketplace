package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const jsAction = "<< /S /JavaScript /JS (app.alert(1)) >>"

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("open-action.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": jsAction}})
	// A document-level script: name tree under /Names /JavaScript, its value
	// an indirect JavaScript action. (pdfcpu's validation drops an empty
	// name tree, so the tree always comes with an action.)
	write("names-tree.pdf", pdfxtest.Doc{
		Catalog: map[string]string{"/Names": "<< /JavaScript << /Names [(sample) {{obj 0}}] >> >>"},
		Objects: []string{jsAction},
	})
	write("annotation-aa.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Annots: []string{
		"<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] /AA << /E " + jsAction + " >> >>",
	}}}})
	write("next-chain.pdf", pdfxtest.Doc{Catalog: map[string]string{
		"/OpenAction": "<< /S /URI /URI (https://example.org) /Next " + jsAction + " >>",
	}})
	write("orphan.pdf", pdfxtest.Doc{Objects: []string{jsAction}})
	write("uri-only.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "<< /S /URI /URI (https://example.org) >>"}})
	write("object-stream.pdf", pdfxtest.Doc{Objects: []string{jsAction}, ObjStream: true})
}
