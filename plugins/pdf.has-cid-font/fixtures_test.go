package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	content = "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	type1   = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"
	// type0 is a composite font with one CIDFontType2 descendant, the font
	// program itself left out (not embedded).
	type0 = "<< /Type /Font /Subtype /Type0 /BaseFont /Sample /Encoding /Identity-H " +
		"/DescendantFonts [<< /Type /Font /Subtype /CIDFontType2 /BaseFont /Sample " +
		"/CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> " +
		"/FontDescriptor << /Type /FontDescriptor /FontName /Sample /Flags 32 " +
		"/FontBBox [0 0 1000 1000] /ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 700 /StemV 80 >> >>] >>"
)

func page(font string) []pdfxtest.Page {
	return []pdfxtest.Page{{Content: content, Resources: "<< /Font << /F1 " + font + " >> >>"}}
}

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("type1.pdf", pdfxtest.Doc{Pages: page(type1)})
	write("no-font.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Content: "0 0 10 10 re f"}}})
	write("type0.pdf", pdfxtest.Doc{Pages: page(type0)})
	write("type0-indirect.pdf", pdfxtest.Doc{Pages: page("{{obj 0}}"), Objects: []string{type0}})
	write("object-stream.pdf", pdfxtest.Doc{Pages: page("{{obj 0}}"), Objects: []string{type0}, ObjStream: true})
	// A Type0 font no page uses still makes the file carry a CID font.
	write("orphan.pdf", pdfxtest.Doc{Pages: page(type1), Objects: []string{type0}})
}
