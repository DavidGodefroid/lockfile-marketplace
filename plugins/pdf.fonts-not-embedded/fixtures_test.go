package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func withFont(objs []string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 {{obj 0}} >> >>"}},
		Objects: objs,
	})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "helvetica.pdf", withFont([]string{pdfxtest.Helvetica}))
	pdfxtest.Raw(t, "testdata", "embedded-ttf.pdf", withFont(pdfxtest.EmbeddedTrueType(0)))
	pdfxtest.Raw(t, "testdata", "unembedded-ttf.pdf", withFont(pdfxtest.UnembeddedTrueType(0)))
	pdfxtest.Raw(t, "testdata", "type0-embedded.pdf", withFont(pdfxtest.Type0(0, true)))
	pdfxtest.Raw(t, "testdata", "type0-unembedded.pdf", withFont(pdfxtest.Type0(0, false)))
	pdfxtest.Raw(t, "testdata", "no-fonts.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{Content: "0 0 10 10 re f"}}}))
}
