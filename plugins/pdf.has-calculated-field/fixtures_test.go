package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		d.Pages = page
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	form := func(extra string) map[string]string {
		return map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}]" + extra + " >>"}
	}
	field := "<< /FT /Tx /T (sample) /V (sample) /DA (/Helv 0 Tf 0 g) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"
	write("plain.pdf", pdfxtest.Doc{})
	write("acroform.pdf", pdfxtest.Doc{Catalog: form(""), Objects: []string{field}})
	write("empty-co.pdf", pdfxtest.Doc{Catalog: form(" /CO []"), Objects: []string{field}})
	write("co.pdf", pdfxtest.Doc{Catalog: form(" /CO [{{obj 0}}]"), Objects: []string{field}})
	// /CO as an indirect array.
	write("indirect-co.pdf", pdfxtest.Doc{Catalog: form(" /CO {{obj 1}}"), Objects: []string{field, "[{{obj 0}}]"}})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form(" /CO [{{obj 0}}]"),
		Objects:   []string{field},
		ObjStream: true,
	})
	// /CO outside the AcroForm dictionary does not count.
	write("stray-co.pdf", pdfxtest.Doc{
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] >>", "/CO": "[{{obj 0}}]"},
		Objects: []string{field},
	})
}
