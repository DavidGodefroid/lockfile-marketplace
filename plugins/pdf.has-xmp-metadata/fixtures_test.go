package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}))
	pdfxtest.Raw(t, "testdata", "xmp.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Metadata: pdfxtest.XMP(pdfxtest.NoPDFA)}))
	pdfxtest.Raw(t, "testdata", "xmp-on-page-only.pdf", pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Resources: "<< /XObject << >> >>"}},
		Objects: []string{pdfxtest.Stream("<< /Type /Metadata /Subtype /XML >>", pdfxtest.XMP(pdfxtest.NoPDFA))},
	}))
}
