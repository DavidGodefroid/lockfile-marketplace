package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func withFont(objs ...string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 {{obj 0}} >> >>"}},
		Objects: objs,
	})
}

func type0(toUnicode bool) []string {
	objs := pdfxtest.Type0(0, true)
	if toUnicode {
		objs[0] = objs[0][:len(objs[0])-2] + "/ToUnicode {{obj 4}} >>"
		objs = append(objs, pdfxtest.Stream("<< >>", []byte("sample")))
	}
	return objs
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-fonts.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{Content: "0 0 10 10 re f"}}}))
	pdfxtest.Raw(t, "testdata", "winansi.pdf", withFont("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>"))
	pdfxtest.Raw(t, "testdata", "base-encoding.pdf", withFont("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding {{obj 1}} >>",
		"<< /Type /Encoding /BaseEncoding /MacRomanEncoding /Differences [32 /space] >>"))
	pdfxtest.Raw(t, "testdata", "simple-tounicode.pdf", withFont("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /ToUnicode {{obj 1}} >>",
		pdfxtest.Stream("<< >>", []byte("sample"))))
	pdfxtest.Raw(t, "testdata", "type0-tounicode.pdf", withFont(type0(true)...))
	pdfxtest.Raw(t, "testdata", "no-encoding.pdf", withFont(pdfxtest.Helvetica))
	pdfxtest.Raw(t, "testdata", "differences-only.pdf", withFont("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding << /Differences [32 /space] >> >>"))
	pdfxtest.Raw(t, "testdata", "type3.pdf", withFont(pdfxtest.Type3))
	pdfxtest.Raw(t, "testdata", "type0-no-tounicode.pdf", withFont(type0(false)...))
	pdfxtest.Raw(t, "testdata", "unused-font.pdf", pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "0 0 10 10 re f"}},
		Objects: type0(false),
	}))
}
