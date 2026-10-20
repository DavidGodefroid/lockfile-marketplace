package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func pages(n int) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: make([]pdfxtest.Page, n)})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "one-page.pdf", pages(1))
	pdfxtest.Raw(t, "testdata", "two-pages.pdf", pages(2))
	pdfxtest.Raw(t, "testdata", "five-pages.pdf", pages(5))
}
