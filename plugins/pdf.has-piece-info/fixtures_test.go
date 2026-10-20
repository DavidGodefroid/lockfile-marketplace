package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// pieceInfo is a page-piece dictionary with one sample application entry.
const pieceInfo = "<< /SampleApp << /LastModified (D:20260101000000Z) /Private << /Sample 1 >> >> >>"

func writeFixtures(t *testing.T) {
	content := "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	res := "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"
	page := func(entries map[string]string) pdfxtest.Page {
		return pdfxtest.Page{Content: content, Resources: res, Entries: entries}
	}
	piece := map[string]string{"/PieceInfo": pieceInfo}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{page(nil)}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("catalog.pdf", pdfxtest.Doc{Catalog: piece})
	write("catalog-indirect.pdf", pdfxtest.Doc{Catalog: map[string]string{"/PieceInfo": "{{obj 0}}"}, Objects: []string{pieceInfo}})
	write("page.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(piece)}})
	write("second-page.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(nil), page(piece)}})
	write("object-stream.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(piece)}, ObjStream: true})
	write("null.pdf", pdfxtest.Doc{Catalog: map[string]string{"/PieceInfo": "null"}, Pages: []pdfxtest.Page{page(map[string]string{"/PieceInfo": "null"})}})
	write("unlisted.pdf", pdfxtest.Doc{Objects: []string{"<< /Type /Page /PieceInfo " + pieceInfo + " >>"}})
}
