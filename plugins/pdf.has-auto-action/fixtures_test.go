package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	uri  = "<< /Type /Action /S /URI /URI (https://example.org) >>"
	goTo = "<< /Type /Action /S /GoTo /D [3 0 R /Fit] >>"
	link = "<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] /A " + uri
)

func writeFixtures(t *testing.T) {
	text := pdfxtest.Page{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{text}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	page := func(p pdfxtest.Page) []pdfxtest.Page {
		p.Content, p.Resources = text.Content, text.Resources
		return []pdfxtest.Page{p}
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("open-action.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": uri}})
	write("open-action-goto-next.pdf", pdfxtest.Doc{Catalog: map[string]string{
		"/OpenAction": "<< /Type /Action /S /GoTo /D [3 0 R /Fit] /Next " + uri + " >>",
	}})
	write("open-action-indirect.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "{{obj 0}}"}, Objects: []string{uri}})
	write("object-stream.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "{{obj 0}}"}, Objects: []string{uri}, ObjStream: true})
	write("catalog-aa.pdf", pdfxtest.Doc{Catalog: map[string]string{"/AA": "<< /WC " + uri + " >>"}})
	write("page-aa.pdf", pdfxtest.Doc{Pages: page(pdfxtest.Page{Entries: map[string]string{"/AA": "<< /O " + uri + " >>"}})})
	write("annot-aa-po.pdf", pdfxtest.Doc{Pages: page(pdfxtest.Page{Annots: []string{link + " /AA << /PO " + uri + " >> >>"}})})
	write("annot-aa-pv.pdf", pdfxtest.Doc{Pages: page(pdfxtest.Page{Annots: []string{link + " /AA << /PV " + uri + " >> >>"}})})

	// A plain go-to, a destination, and gesture-driven triggers do not
	// run by themselves.
	write("open-action-goto.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": goTo}})
	write("open-action-dest.pdf", pdfxtest.Doc{Catalog: map[string]string{"/OpenAction": "[3 0 R /Fit]"}})
	write("empty-aa.pdf", pdfxtest.Doc{Catalog: map[string]string{"/AA": "<< >>"}})
	write("annot-aa-gesture.pdf", pdfxtest.Doc{Pages: page(pdfxtest.Page{Annots: []string{link + " /AA << /E " + uri + " /D " + uri + " >> >>"}})})
	write("link-only.pdf", pdfxtest.Doc{Pages: page(pdfxtest.Page{Annots: []string{link + " >>"}})})
}
