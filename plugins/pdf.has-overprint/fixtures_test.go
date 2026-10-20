package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	const content = "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	font := "/Font << /F1 " + pdfxtest.Helvetica + " >>"
	page := func(gs string) []pdfxtest.Page {
		return []pdfxtest.Page{{Content: "/GS1 gs " + content, Resources: "<< " + font + " /ExtGState << /GS1 " + gs + " >> >>"}}
	}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{{Content: content, Resources: "<< " + font + " >>"}}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	// Overprint keys present but off, and an ExtGState without them.
	write("overprint-off.pdf", pdfxtest.Doc{
		Pages:   page("<< /Type /ExtGState /OP false /op false /OPM 1 >>"),
		Objects: []string{"<< /Type /ExtGState /CA 0.5 /ca 0.5 >>"},
	})
	write("stroke-overprint.pdf", pdfxtest.Doc{
		Pages: page("<< /Type /ExtGState /OP true >>"),
	})
	write("fill-overprint.pdf", pdfxtest.Doc{
		Pages:   page("{{obj 0}}"),
		Objects: []string{"<< /op true /OP false >>"},
	})
	write("overprint-indirect.pdf", pdfxtest.Doc{
		Pages:     page("{{obj 0}}"),
		Objects:   []string{"<< /Type /ExtGState /OP {{obj 1}} >>", "true"},
		ObjStream: true,
	})
}
