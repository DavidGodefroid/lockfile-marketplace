package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	gotoR = "<< /Type /Action /S /GoToR /F (sample.pdf) /D [0 /Fit] >>"
	gotoE = "<< /Type /Action /S /GoToE /D [0 /Fit] /T << /R /C /N (sample) >> >>"
)

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("open-action-gotor.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": gotoR}})
	write("link-gotoe.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Annots: []string{
		"<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] /A " + gotoE + " >>",
	}}}})
	write("next-chain.pdf", pdfxtest.Doc{Catalog: map[string]string{
		"/OpenAction": "<< /S /URI /URI (https://example.org) /Next " + gotoR + " >>",
	}})
	write("orphan.pdf", pdfxtest.Doc{Objects: []string{gotoE}})
	write("object-stream.pdf", pdfxtest.Doc{Objects: []string{gotoR}, ObjStream: true})
	// A local go-to stays inside the document: not remote.
	write("local-goto.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "<< /S /GoTo /D (sample) >>"}})
	// /S also names other things: a dict typed as something else is not
	// an action.
	write("not-an-action.pdf", pdfxtest.Doc{Objects: []string{"<< /Type /Sample /S /GoToR >>"}})
}
