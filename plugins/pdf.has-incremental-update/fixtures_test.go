package main

import (
	"bytes"
	"strconv"
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

func update(b []byte) []byte {
	return pdfxtest.AppendUpdate(b, []string{"<< /Producer (sample) >>"}, "/Info {{obj 0}}")
}

// selfLoop makes the only section's trailer /Prev point to itself.
func selfLoop(b []byte) []byte {
	off := bytes.LastIndex(b, []byte("xref\n0 "))
	return bytes.Replace(b, []byte("/Root 1 0 R >>\nstartxref"),
		[]byte("/Root 1 0 R /Prev "+strconv.Itoa(off)+" >>\nstartxref"), 1)
}

func garbageStartXRef(b []byte) []byte {
	i := bytes.LastIndex(b, []byte("startxref\n"))
	return append(bytes.Clone(b[:i]), "startxref\n12\n%%EOF\n"...)
}

func writeFixtures(t *testing.T) {
	xs := func(d *pdfxtest.Doc) { d.XRefStream = true }
	pdfxtest.Raw(t, "testdata", "fresh.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "one-update.pdf", update(doc(nil)))
	pdfxtest.Raw(t, "testdata", "two-updates.pdf", update(update(doc(nil))))
	pdfxtest.Raw(t, "testdata", "xref-stream.pdf", doc(xs))
	pdfxtest.Raw(t, "testdata", "xref-stream-update.pdf", update(doc(xs)))
	pdfxtest.Raw(t, "testdata", "linearized.pdf", doc(func(d *pdfxtest.Doc) { d.Linearized = true }))
	pdfxtest.Raw(t, "testdata", "linearized-update.pdf", update(doc(func(d *pdfxtest.Doc) { d.Linearized = true })))
	pdfxtest.Raw(t, "testdata", "hybrid.pdf", doc(func(d *pdfxtest.Doc) { d.Hybrid = true }))
}
