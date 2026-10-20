package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func pages(ps ...pdfxtest.Page) []byte { return pdfxtest.Write(pdfxtest.Doc{Pages: ps}) }

func box(b [4]float64) *[4]float64 { return &b }

func writeFixtures(t *testing.T) {
	landscape := [4]float64{0, 0, 841.890, 595.276}
	halfMMOff := [4]float64{0, 0, 595.276 + 1.417, 841.890 - 1.417}
	a4 := pdfxtest.Page{}
	pdfxtest.Raw(t, "testdata", "one-page.pdf", pages(a4))
	pdfxtest.Raw(t, "testdata", "a4-and-a4-landscape.pdf", pages(a4, pdfxtest.Page{MediaBox: box(landscape)}))
	pdfxtest.Raw(t, "testdata", "a4-and-half-mm-off.pdf", pages(a4, pdfxtest.Page{MediaBox: box(halfMMOff)}))
	pdfxtest.Raw(t, "testdata", "a4-and-letter.pdf", pages(a4, pdfxtest.Page{MediaBox: box(pdfxtest.Letter)}))
	pdfxtest.Raw(t, "testdata", "a4-then-a5-crop.pdf", pages(a4, a4, a4, pdfxtest.Page{CropBox: box(pdfxtest.A5)}))
}
