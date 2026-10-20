package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const submit = "<< /Type /Action /S /SubmitForm /F << /FS /URL /F (https://example.org/sample) >> >>"

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("widget.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Annots: []string{
		"<< /Type /Annot /Subtype /Widget /FT /Btn /Ff 65536 /T (sample) /Rect [0 0 10 10] /A " + submit + " >>",
	}}}})
	write("open-action.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": submit}})
	write("next-chain.pdf", pdfxtest.Doc{Catalog: map[string]string{
		"/OpenAction": "<< /S /URI /URI (https://example.org) /Next " + submit + " >>",
	}})
	write("orphan.pdf", pdfxtest.Doc{Objects: []string{submit}})
	write("object-stream.pdf", pdfxtest.Doc{Objects: []string{submit}, ObjStream: true})
	// A reset-form button is a form action too, but sends nothing.
	write("reset-only.pdf", pdfxtest.Doc{Objects: []string{"<< /Type /Action /S /ResetForm >>"}})
	// /S also names other things: a dict typed as something else is not
	// an action.
	write("not-an-action.pdf", pdfxtest.Doc{Objects: []string{"<< /Type /Sample /S /SubmitForm >>"}})
}
