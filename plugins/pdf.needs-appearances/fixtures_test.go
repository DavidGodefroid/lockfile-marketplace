package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const field = "<< /FT /Tx /T (sample) /V (sample) /DA (/Helv 0 Tf 0 g) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"

var page = []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}

// formDoc is a one-field form whose AcroForm dictionary ends with extra.
func formDoc(extra string, objects ...string) pdfxtest.Doc {
	return pdfxtest.Doc{
		Pages:   page,
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}]" + extra + " >>"},
		Objects: append([]string{field}, objects...),
	}
}

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{Pages: page})
	write("acroform.pdf", formDoc(""))
	write("false.pdf", formDoc(" /NeedAppearances false"))
	write("true.pdf", formDoc(" /NeedAppearances true"))
	write("indirect.pdf", formDoc(" /NeedAppearances {{obj 1}}", "true"))
	stream := formDoc(" /NeedAppearances true")
	stream.ObjStream = true
	write("object-stream.pdf", stream)
	// /NeedAppearances outside the AcroForm dictionary does not count.
	stray := formDoc("")
	stray.Catalog["/NeedAppearances"] = "true"
	write("stray.pdf", stray)
}
