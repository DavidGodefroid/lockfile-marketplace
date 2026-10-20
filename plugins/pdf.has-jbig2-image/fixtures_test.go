package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const image = "/Type /XObject /Subtype /Image /Width 1 /Height 1 /BitsPerComponent 1 /ColorSpace /DeviceGray"

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "q 1 0 0 1 0 0 cm /Im0 Do Q", Resources: "<< /XObject << /Im0 {{obj 0}} >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}})
	write("flate-image.pdf", pdfxtest.Doc{Objects: []string{pdfxtest.Stream("<< "+image+" /Filter /FlateDecode >>", []byte("sample"))}})
	write("jbig2-name.pdf", pdfxtest.Doc{Objects: []string{pdfxtest.Stream("<< "+image+" /Filter /JBIG2Decode >>", []byte("sample"))}})
	write("jbig2-array.pdf", pdfxtest.Doc{Objects: []string{pdfxtest.Stream("<< "+image+" /Filter [/ASCIIHexDecode /JBIG2Decode] >>", []byte("sample"))}})
	write("jbig2-indirect.pdf", pdfxtest.Doc{Objects: []string{
		pdfxtest.Stream("<< "+image+" /Filter {{obj 1}} >>", []byte("sample")),
		"[/JBIG2Decode]",
	}})
	write("jbig2-orphan.pdf", pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}},
		Objects: []string{pdfxtest.Stream("<< "+image+" /Filter /JBIG2Decode >>", []byte("sample"))},
	})
	// The filter name outside a stream dict is not a stream filter.
	write("jbig2-not-stream.pdf", pdfxtest.Doc{Objects: []string{pdfxtest.Stream("<< "+image+" >>", []byte("sample")), "<< /Filter /JBIG2Decode >>"}})
}
