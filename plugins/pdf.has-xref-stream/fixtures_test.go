package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(mod func(*pdfxtest.Doc)) []byte {
	d := pdfxtest.Pages("text", "blank")
	if mod != nil {
		mod(&d)
	}
	return pdfxtest.Write(d)
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "classic.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "classic-updated.pdf",
		pdfxtest.AppendUpdate(doc(nil), []string{"<< /Producer (sample) >>"}, "/Info {{obj 0}}"))
	pdfxtest.Raw(t, "testdata", "hybrid.pdf", doc(func(d *pdfxtest.Doc) { d.Hybrid = true }))
	pdfxtest.Raw(t, "testdata", "xref-stream.pdf", doc(func(d *pdfxtest.Doc) { d.XRefStream = true }))
	pdfxtest.Raw(t, "testdata", "object-stream.pdf", doc(func(d *pdfxtest.Doc) { d.ObjStream = true }))
	pdfxtest.Raw(t, "testdata", "linearized-xref-stream.pdf", doc(func(d *pdfxtest.Doc) { d.XRefStream, d.Linearized = true, true }))
}
