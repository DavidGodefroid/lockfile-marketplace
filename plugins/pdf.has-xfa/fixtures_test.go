package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		d.Pages = page
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	form := func(extra string) map[string]string {
		return map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}]" + extra + " >>"}
	}
	field := "<< /FT /Tx /T (sample) /V (sample) /DA (/Helv 0 Tf 0 g) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"
	xdp := pdfxtest.Stream("<< >>", []byte("<xdp:xdp xmlns:xdp=\"http://ns.adobe.com/xdp/\"></xdp:xdp>"))
	write("plain.pdf", pdfxtest.Doc{})
	write("acroform.pdf", pdfxtest.Doc{Catalog: form(""), Objects: []string{field}})
	write("xfa-stream.pdf", pdfxtest.Doc{Catalog: form(" /XFA {{obj 1}}"), Objects: []string{field, xdp}})
	// The XML split into packets: [name stream name stream ...].
	write("xfa-packets.pdf", pdfxtest.Doc{
		Catalog: form(" /XFA [(preamble) {{obj 1}} (postamble) {{obj 2}}]"),
		Objects: []string{
			field,
			pdfxtest.Stream("<< >>", []byte("<xdp:xdp xmlns:xdp=\"http://ns.adobe.com/xdp/\">")),
			pdfxtest.Stream("<< >>", []byte("</xdp:xdp>")),
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form(" /XFA {{obj 1}}"),
		Objects:   []string{field, xdp},
		ObjStream: true,
	})
	// /XFA outside the AcroForm dictionary does not count.
	write("stray-xfa.pdf", pdfxtest.Doc{
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] >>", "/XFA": "{{obj 1}}"},
		Objects: []string{field, xdp},
	})
}
