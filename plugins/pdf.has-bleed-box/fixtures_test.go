package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func bleed(box string) map[string]string { return map[string]string{"/BleedBox": box} }

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) { pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d)) }
	write("no-bleed-box.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}}})
	write("trim-box-only.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Entries: map[string]string{"/TrimBox": "[10 10 585 832]"}}}})
	write("bleed-box-page-1.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Entries: bleed("[0 0 595 842]")}}})
	write("bleed-box-page-2-of-3.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {Entries: bleed("[5 5 590 837]")}, {}}})
	write("bleed-box-indirect.pdf", pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Entries: bleed("{{obj 0}}")}},
		Objects: []string{"[0 0 595 842]"},
	})
}
