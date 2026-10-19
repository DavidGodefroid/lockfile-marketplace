package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func rot(v int) *int { return &v }

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) { pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d)) }
	write("no-rotate.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}}})
	write("rotate-0.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Rotate: rot(0)}}})
	write("rotate-90-page-2-of-3.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {Rotate: rot(90)}, {}}})
	write("rotate-minus-90.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Rotate: rot(-90)}}})
	write("rotate-360.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Rotate: rot(360)}}})
	write("rotate-90-on-pages-node.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, PagesRotate: rot(90)})
}
