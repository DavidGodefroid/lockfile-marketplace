package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func pages(ps ...pdfxtest.Page) []byte { return pdfxtest.Write(pdfxtest.Doc{Pages: ps}) }

func box(b [4]float64) *[4]float64 { return &b }

func writeFixtures(t *testing.T) {
	a3 := [4]float64{0, 0, 841.890, 1190.551}
	a3Landscape := [4]float64{0, 0, 1190.551, 841.890}
	a3PlusHalfMM := [4]float64{0, 0, 841.890 + 1.417, 1190.551 + 1.417}
	a2 := [4]float64{0, 0, 1190.551, 1683.780}
	square300mm := [4]float64{0, 0, 850.394, 850.394}
	strip100x430mm := [4]float64{0, 0, 283.465, 1218.898}
	a4 := [4]float64{0, 0, 595.276, 841.890}
	pdfxtest.Raw(t, "testdata", "a4.pdf", pages(pdfxtest.Page{}))
	pdfxtest.Raw(t, "testdata", "a3-portrait.pdf", pages(pdfxtest.Page{MediaBox: box(a3)}))
	pdfxtest.Raw(t, "testdata", "a3-landscape.pdf", pages(pdfxtest.Page{MediaBox: box(a3Landscape)}))
	pdfxtest.Raw(t, "testdata", "a3-half-mm-over.pdf", pages(pdfxtest.Page{MediaBox: box(a3PlusHalfMM)}))
	pdfxtest.Raw(t, "testdata", "a2.pdf", pages(pdfxtest.Page{MediaBox: box(a2)}))
	pdfxtest.Raw(t, "testdata", "square-300mm.pdf", pages(pdfxtest.Page{MediaBox: box(square300mm)}))
	pdfxtest.Raw(t, "testdata", "strip-100x430mm.pdf", pages(pdfxtest.Page{MediaBox: box(strip100x430mm)}))
	pdfxtest.Raw(t, "testdata", "a4-then-a2.pdf", pages(pdfxtest.Page{}, pdfxtest.Page{MediaBox: box(a2)}))
	pdfxtest.Raw(t, "testdata", "a2-media-a4-crop.pdf", pages(pdfxtest.Page{MediaBox: box(a2), CropBox: box(a4)}))
	pdfxtest.Raw(t, "testdata", "userunit-2-a4-box.pdf", pages(pdfxtest.Page{MediaBox: box(a4), UserUnit: 2}))
}
