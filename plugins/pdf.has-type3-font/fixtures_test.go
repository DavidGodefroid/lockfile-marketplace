package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// type3 is a Type 3 font whose one glyph, a square, is drawn by the
// content stream {{obj 1}}.
const type3 = "<< /Type /Font /Subtype /Type3 /FontBBox [0 0 750 750] /FontMatrix [0.001 0 0 0.001 0 0] " +
	"/CharProcs << /square {{obj 1}} >> /Encoding << /Type /Encoding /Differences [97 /square] >> " +
	"/FirstChar 97 /LastChar 97 /Widths [1000] /Resources << >> >>"

var glyph = pdfxtest.Stream("<< >>", []byte("1000 0 0 0 750 750 d1 0 0 750 750 re f"))

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("cid-font.pdf", pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td <0001> Tj ET", Resources: "<< /Font << /F1 {{obj 0}} >> >>"}},
		Objects: pdfxtest.Type0(0, true),
	})
	write("type3-page.pdf", pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (a) Tj ET", Resources: "<< /Font << /F1 {{obj 0}} >> >>"}},
		Objects: []string{type3, glyph},
	})
	write("orphan-type3.pdf", pdfxtest.Doc{Objects: []string{type3, glyph}})
	write("object-stream.pdf", pdfxtest.Doc{
		Pages:     []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (a) Tj ET", Resources: "<< /Font << /F1 {{obj 0}} >> >>"}},
		Objects:   []string{type3, glyph},
		ObjStream: true,
	})
	// /Subtype /Type3 on a dictionary that is not a font.
	write("not-a-font.pdf", pdfxtest.Doc{Objects: []string{"<< /Type /Sample /Subtype /Type3 >>"}})
}
