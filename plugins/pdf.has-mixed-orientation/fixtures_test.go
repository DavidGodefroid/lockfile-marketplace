package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func box(b [4]float64) *[4]float64 { return &b }

func rot(v int) *int { return &v }

var (
	portrait  = pdfxtest.Page{}
	landscape = pdfxtest.Page{MediaBox: box([4]float64{0, 0, 841.890, 595.276})}
	square    = pdfxtest.Page{MediaBox: box([4]float64{0, 0, 500, 500})}
)

func pages(ps ...pdfxtest.Page) []byte { return pdfxtest.Write(pdfxtest.Doc{Pages: ps}) }

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "all-portrait.pdf", pages(portrait, portrait))
	pdfxtest.Raw(t, "testdata", "portrait-and-landscape.pdf", pages(portrait, landscape))
	pdfxtest.Raw(t, "testdata", "portrait-and-rotated-portrait.pdf", pages(portrait, pdfxtest.Page{Rotate: rot(90)}))
	pdfxtest.Raw(t, "testdata", "landscape-all-rotated.pdf",
		pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{landscape, landscape}, PagesRotate: rot(270)}))
	pdfxtest.Raw(t, "testdata", "portrait-and-square.pdf", pages(portrait, square))
	pdfxtest.Raw(t, "testdata", "portrait-and-upside-down.pdf", pages(portrait, pdfxtest.Page{Rotate: rot(180)}))
}
