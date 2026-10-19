package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func pages(ps ...pdfxtest.Page) []byte { return pdfxtest.Write(pdfxtest.Doc{Pages: ps}) }

func box(b [4]float64) *[4]float64 { return &b }

func writeFixtures(t *testing.T) {
	landscape := [4]float64{0, 0, 841.890, 595.276}
	plusHalfMM := [4]float64{0, 0, 595.276 + 1.417, 841.890 - 1.417}
	half := [4]float64{0, 0, 297.638, 420.945}
	pdfxtest.Raw(t, "testdata", "a4-portrait.pdf", pages(pdfxtest.Page{}))
	pdfxtest.Raw(t, "testdata", "a4-landscape.pdf", pages(pdfxtest.Page{MediaBox: box(landscape)}))
	pdfxtest.Raw(t, "testdata", "a4-half-mm-off.pdf", pages(pdfxtest.Page{MediaBox: box(plusHalfMM)}))
	pdfxtest.Raw(t, "testdata", "letter.pdf", pages(pdfxtest.Page{MediaBox: box(pdfxtest.Letter)}))
	pdfxtest.Raw(t, "testdata", "a4-media-a5-crop.pdf", pages(pdfxtest.Page{CropBox: box(pdfxtest.A5)}))
	pdfxtest.Raw(t, "testdata", "a4-then-letter.pdf", pages(pdfxtest.Page{}, pdfxtest.Page{MediaBox: box(pdfxtest.Letter)}))
	pdfxtest.Raw(t, "testdata", "userunit-2-half-box.pdf", pages(pdfxtest.Page{MediaBox: box(half), UserUnit: 2}))
}
