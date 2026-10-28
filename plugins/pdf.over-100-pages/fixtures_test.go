package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func pages(n int) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: make([]pdfxtest.Page, n)})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "zero-pages.pdf", pages(0))
	pdfxtest.Raw(t, "testdata", "one-page.pdf", pages(1))
	pdfxtest.Raw(t, "testdata", "100-pages.pdf", pages(100))
	pdfxtest.Raw(t, "testdata", "101-pages.pdf", pages(101))
	pdfxtest.Raw(t, "testdata", "250-pages.pdf", pages(250))
}
