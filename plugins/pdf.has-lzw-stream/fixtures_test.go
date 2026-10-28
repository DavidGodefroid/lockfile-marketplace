package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	// Other filters, and LZWDecode only as a name outside a stream dict.
	write("other-filters.pdf", pdfxtest.Doc{Objects: []string{
		pdfxtest.Stream("<< /Filter /ASCIIHexDecode >>", []byte("73616d706c65>")),
		pdfxtest.Stream("<< /Filter [/ASCIIHexDecode /ASCII85Decode] >>", []byte("sample")),
		"<< /Filter /LZWDecode >>",
	}})
	write("lzw-name.pdf", pdfxtest.Doc{Objects: []string{
		pdfxtest.Stream("<< /Filter /LZWDecode >>", []byte("sample")),
	}})
	write("lzw-array.pdf", pdfxtest.Doc{Objects: []string{
		pdfxtest.Stream("<< /Filter [/ASCIIHexDecode /LZWDecode] >>", []byte("sample")),
	}})
	write("lzw-indirect.pdf", pdfxtest.Doc{Objects: []string{
		pdfxtest.Stream("<< /Filter {{obj 1}} >>", []byte("sample")),
		"[/LZWDecode]",
	}, ObjStream: true})
}
