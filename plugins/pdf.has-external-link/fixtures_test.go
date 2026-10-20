package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const uri = "<< /S /URI /URI (https://example.org) >>"

func link(action string) []pdfxtest.Page {
	return []pdfxtest.Page{{Annots: []string{"<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] /A " + action + " >>"}}}
}

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("uri-link.pdf", pdfxtest.Doc{Pages: link(uri)})
	write("remote-goto.pdf", pdfxtest.Doc{Pages: link("<< /S /GoToR /F (sample.pdf) /D [0 /Fit] >>")})
	write("next-chain.pdf", pdfxtest.Doc{Catalog: map[string]string{
		"/OpenAction": "<< /S /GoTo /D [3 0 R /Fit] /Next " + uri + " >>",
	}})
	write("orphan.pdf", pdfxtest.Doc{Objects: []string{uri}})
	write("object-stream.pdf", pdfxtest.Doc{Objects: []string{uri}, ObjStream: true})
	write("internal-link.pdf", pdfxtest.Doc{Pages: link("<< /S /GoTo /D [3 0 R /Fit] >>")})
	write("launch-only.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "<< /S /Launch /F (sample.exe) >>"}})
}
