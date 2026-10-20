package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// thumb is a 1x1 grey thumbnail image stream, written as extra object 0.
var thumb = pdfxtest.Stream("<< /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0x80})

func withThumb() pdfxtest.Page {
	return pdfxtest.Page{Entries: map[string]string{"/Thumb": "{{obj 0}}"}}
}

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) { pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d)) }
	write("no-thumb.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}}})
	write("thumb-page-1.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{withThumb()}, Objects: []string{thumb}})
	write("thumb-page-3-of-3.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}, withThumb()}, Objects: []string{thumb}})
	write("thumb-unreferenced.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{thumb}})
}
