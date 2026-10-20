package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-page-mode.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "fullscreen.pdf", doc(map[string]string{"/PageMode": "/FullScreen"}))
	pdfxtest.Raw(t, "testdata", "use-outlines.pdf", doc(map[string]string{"/PageMode": "/UseOutlines"}))
	pdfxtest.Raw(t, "testdata", "use-none.pdf", doc(map[string]string{"/PageMode": "/UseNone"}))
}
