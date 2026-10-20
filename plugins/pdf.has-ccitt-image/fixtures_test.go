package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// image is a 1x1 bilevel image XObject compressed with filter; its data is
// never decoded.
func image(filter string) string {
	return pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 1 /Filter "+filter+" >>", []byte("sample"))
}

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	drawn := []pdfxtest.Page{{Content: "q 10 0 0 10 72 720 cm /Im0 Do Q", Resources: "<< /XObject << /Im0 {{obj 0}} >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("ccitt-image.pdf", pdfxtest.Doc{Pages: drawn, Objects: []string{image("/CCITTFaxDecode")}})
	write("filter-array.pdf", pdfxtest.Doc{Pages: drawn, Objects: []string{image("[/FlateDecode /CCITTFaxDecode]")}})
	write("indirect-filter.pdf", pdfxtest.Doc{Pages: drawn, Objects: []string{image("{{obj 1}}"), "/CCITTFaxDecode"}})
	write("orphan-stream.pdf", pdfxtest.Doc{Objects: []string{image("/CCITTFaxDecode")}})
	write("runlength-image.pdf", pdfxtest.Doc{Pages: drawn, Objects: []string{image("/RunLengthDecode")}})
	write("flate-array.pdf", pdfxtest.Doc{Pages: drawn, Objects: []string{image("[/ASCIIHexDecode /FlateDecode]")}})
	// The filter name in a non-stream dict does not count.
	write("name-in-dict.pdf", pdfxtest.Doc{Objects: []string{"<< /Filter /CCITTFaxDecode >>"}})
}
