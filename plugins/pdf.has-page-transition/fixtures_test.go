package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// trans is a dissolve transition dictionary, written as extra object 0.
const trans = "<< /Type /Trans /S /Dissolve /D 1 >>"

func page(entries map[string]string) pdfxtest.Page { return pdfxtest.Page{Entries: entries} }

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) { pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d)) }
	write("no-transition.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}}})
	write("trans-page-1.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(map[string]string{"/Trans": trans})}})
	write("trans-ref-page-3-of-3.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}, page(map[string]string{"/Trans": "{{obj 0}}"})}, Objects: []string{trans}})
	write("dur-page-2-of-2.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, page(map[string]string{"/Dur": "5"})}})
	write("trans-null.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(map[string]string{"/Trans": "null"})}})
	write("trans-unreferenced.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{trans}})
}
