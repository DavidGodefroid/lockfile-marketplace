package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// xobject is an XObject stream with the raw extra entries e.
func xobject(e string) string {
	return pdfxtest.Stream("<< /Type /XObject "+e+" >>", []byte("sample"))
}

func writeFixtures(t *testing.T) {
	text := pdfxtest.Page{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}
	drawn := pdfxtest.Page{Content: "q /X0 Do Q", Resources: "<< /XObject << /X0 {{obj 0}} >> >>"}
	form := "/Subtype /Form /BBox [0 0 10 10]"
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{text}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("form.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{xobject(form)}})
	write("ps-xobject.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{xobject("/Subtype /PS")}})
	write("form-subtype2-ps.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{xobject(form + " /Subtype2 /PS")}})
	write("image-subtype2-ps.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{xobject("/Subtype /Image /Width 1 /Height 1 /BitsPerComponent 8 /ColorSpace /DeviceGray /Subtype2 /PS")}})
	write("orphan-ps.pdf", pdfxtest.Doc{Objects: []string{xobject("/Subtype /PS")}})
	write("dict-ps.pdf", pdfxtest.Doc{Catalog: map[string]string{"/PieceInfo": "{{obj 0}}"}, Objects: []string{"<< /Subtype /PS >>"}})
	write("indirect-subtype.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{drawn}, Objects: []string{xobject("/Subtype {{obj 1}}"), "/PS"}})
}
