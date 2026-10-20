package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const textOp = "BT /F1 12 Tf 72 720 Td (sample) Tj ET"

// formDoc is one page drawing form Fm1, which draws form Fm2 when nested;
// the innermost form's content is inner. Objects: 0 font, 1 Fm1, 2 Fm2.
func formDoc(inner string, nested bool) pdfxtest.Doc {
	res := " /Resources << /Font << /F1 {{obj 0}} >> /XObject << /Fm2 {{obj 2}} >> >>"
	form := func(content string) string {
		return pdfxtest.Stream("<< /Type /XObject /Subtype /Form /BBox [0 0 595 842]"+res+" >>", []byte(content))
	}
	objs := []string{pdfxtest.Helvetica}
	if nested {
		objs = append(objs, form("q /Fm2 Do Q"), form(inner))
	} else {
		objs = append(objs, form(inner), form("0 0 10 10 re f"))
	}
	return pdfxtest.Doc{
		Objects: objs,
		Pages:   []pdfxtest.Page{{Content: "q /Fm1 Do Q", Resources: "<< /XObject << /Fm1 {{obj 1}} >> >>"}},
	}
}

func writeFixtures(t *testing.T) {
	for name, kinds := range map[string][]string{
		"text.pdf":                {"text"},
		"three-text-pages.pdf":    {"text", "text", "text"},
		"invisible-text.pdf":      {"invisible-text"},
		"text-then-invisible.pdf": {"text", "invisible-text"},
		"white-then-text.pdf":     {"white-then-text"},
		"blank.pdf":               {"blank"},
		"image.pdf":               {"image"},
		"text-then-blank.pdf":     {"text", "blank"},
		"blank-then-text.pdf":     {"blank", "text"},
		"text-then-image.pdf":     {"text", "image"},
		"text-white-text.pdf":     {"text", "white", "text"},
	} {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(pdfxtest.Pages(kinds...)))
	}
	// A scanned page with an invisible OCR text layer over its image.
	ocr := pdfxtest.Pages("image")
	ocr.Pages[0].Content += " BT 3 Tr /F1 12 Tf 72 720 Td (sample) Tj ET"
	pdfxtest.Raw(t, "testdata", "image-with-ocr.pdf", pdfxtest.Write(ocr))

	pdfxtest.Raw(t, "testdata", "form-text.pdf", pdfxtest.Write(formDoc(textOp, false)))
	pdfxtest.Raw(t, "testdata", "nested-form-text.pdf", pdfxtest.Write(formDoc(textOp, true)))
	pdfxtest.Raw(t, "testdata", "form-no-text.pdf", pdfxtest.Write(formDoc("0 0 10 10 re f", true)))
	objStm := formDoc(textOp, true)
	objStm.ObjStream = true
	pdfxtest.Raw(t, "testdata", "nested-form-text-objstm.pdf", pdfxtest.Write(objStm))
}
