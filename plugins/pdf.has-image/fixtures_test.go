package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const imageOp = "q 10 0 0 10 0 0 cm /Im1 Do Q"

// formDoc is one page drawing form Fm1, which draws form Fm2 when nested;
// the innermost form's content is inner. Objects: 0 font, 1 image, 2 Fm1,
// 3 Fm2.
func formDoc(inner string, nested bool) pdfxtest.Doc {
	res := " /Resources << /Font << /F1 {{obj 0}} >> /XObject << /Im1 {{obj 1}} /Fm2 {{obj 3}} >> >>"
	objs := []string{
		pdfxtest.Helvetica,
		pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0x80}),
	}
	form := func(content string) string {
		return pdfxtest.Stream("<< /Type /XObject /Subtype /Form /BBox [0 0 595 842]"+res+" >>", []byte(content))
	}
	if nested {
		objs = append(objs, form("q /Fm2 Do Q"), form(inner))
	} else {
		objs = append(objs, form(inner), form("0 0 10 10 re f"))
	}
	return pdfxtest.Doc{
		Objects: objs,
		Pages:   []pdfxtest.Page{{Content: "q /Fm1 Do Q", Resources: "<< /XObject << /Fm1 {{obj 2}} >> >>"}},
	}
}

func writeFixtures(t *testing.T) {
	for name, kinds := range map[string][]string{
		"image.pdf":           {"image"},
		"text-then-image.pdf": {"text", "blank", "image"},
		"blank.pdf":           {"blank"},
		"text.pdf":            {"text"},
		"white-fill.pdf":      {"white"},
		"invisible-text.pdf":  {"invisible-text"},
		"text-and-blank.pdf":  {"text", "blank", "white"},
	} {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(pdfxtest.Pages(kinds...)))
	}
	pdfxtest.Raw(t, "testdata", "form-image.pdf", pdfxtest.Write(formDoc(imageOp, false)))
	pdfxtest.Raw(t, "testdata", "nested-form-image.pdf", pdfxtest.Write(formDoc(imageOp, true)))
	pdfxtest.Raw(t, "testdata", "form-text.pdf", pdfxtest.Write(formDoc("BT /F1 12 Tf 72 720 Td (sample) Tj ET", true)))
	// The resources name Im1 but the page never draws it.
	pdfxtest.Raw(t, "testdata", "unused-image.pdf", pdfxtest.Write(formDoc("0 0 10 10 re f", true)))
	// Do of a name absent from the resources draws nothing.
	pdfxtest.Raw(t, "testdata", "missing-image.pdf", pdfxtest.Write(formDoc("q /Im9 Do Q", true)))
	inline := pdfxtest.Doc{Pages: []pdfxtest.Page{{Content: "q 10 0 0 10 72 600 cm BI /W 1 /H 1 /CS /G /BPC 8 ID \x80 EI Q"}}}
	pdfxtest.Raw(t, "testdata", "inline-image.pdf", pdfxtest.Write(inline))
	objStm := formDoc(imageOp, true)
	objStm.ObjStream = true
	pdfxtest.Raw(t, "testdata", "nested-form-image-objstm.pdf", pdfxtest.Write(objStm))
}
