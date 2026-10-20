package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// image is an image XObject dict with filter f; its data is never decoded.
func image(f string) string {
	return pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /BitsPerComponent 8 /ColorSpace /DeviceGray /Filter "+f+" >>", []byte("sample"))
}

func writeFixtures(t *testing.T) {
	text := pdfxtest.Page{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}
	drawn := pdfxtest.Page{Content: "q 10 0 0 10 72 720 cm /Im0 Do Q", Resources: "<< /XObject << /Im0 {{obj 0}} >> >>"}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{text}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("jpeg-image.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{image("/DCTDecode")}})
	write("jpx-image.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{image("/JPXDecode")}})
	write("jpx-in-chain.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{image("[/FlateDecode /JPXDecode]")}})
	write("flate-chain.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{image("[/ASCIIHexDecode /FlateDecode]")}})
	write("orphan-jpx.pdf", pdfxtest.Doc{Objects: []string{image("/JPXDecode")}})
	write("indirect-filter.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{image("{{obj 1}}"), "/JPXDecode"}})
}
