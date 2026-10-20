package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const launch = "<< /Type /Action /S /Launch /F (sample.exe) >>"

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("open-action.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": launch}})
	write("link.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Annots: []string{
		"<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] /A " + launch + " >>",
	}}}})
	write("next-chain.pdf", pdfxtest.Doc{Catalog: map[string]string{
		"/OpenAction": "<< /S /URI /URI (https://example.org) /Next " + launch + " >>",
	}})
	write("orphan.pdf", pdfxtest.Doc{Objects: []string{launch}})
	write("object-stream.pdf", pdfxtest.Doc{Objects: []string{launch}, ObjStream: true})
	write("uri-only.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "<< /S /URI /URI (https://example.org) >>"}})
	// /S also names other things: a dict typed as something else is not
	// an action.
	write("not-an-action.pdf", pdfxtest.Doc{Objects: []string{"<< /Type /Sample /S /Launch >>"}})
}
