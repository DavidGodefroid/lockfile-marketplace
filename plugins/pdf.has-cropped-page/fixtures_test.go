package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func box(llx, lly, urx, ury float64) *[4]float64 { return &[4]float64{llx, lly, urx, ury} }

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) { pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d)) }
	write("no-crop-box.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {}}})
	write("crop-equals-media.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: &pdfxtest.A4}}})
	write("crop-within-1mm.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: box(2, 2, 593, 840)}}})
	write("crop-larger-than-media.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: box(-10, -10, 700, 900)}}})
	write("crop-a5-page-2-of-3.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {CropBox: &pdfxtest.A5}, {}}})
	write("crop-one-edge-by-2mm.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: box(5.67, 0, 595.276, 841.890)}}})
	write("crop-own-media-box.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{MediaBox: &pdfxtest.Letter, CropBox: box(0, 0, 612, 792)}}})
	write("crop-indirect.pdf", pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Entries: map[string]string{"/CropBox": "{{obj 0}}"}}},
		Objects: []string{"[0 0 300 400]"},
	})
	write("crop-under-1mm-scaled-by-user-unit.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: box(0, 0, 595.276, 840.5), UserUnit: 3}}})
}
