package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func annot(subtype, extra string) string {
	return "<< /Type /Annot /Subtype /" + subtype + " /Rect [72 700 144 720]" + extra + " >>"
}

func writeFixtures(t *testing.T) {
	content := "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	res := "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"
	page := func(annots ...string) pdfxtest.Page {
		return pdfxtest.Page{Content: content, Resources: res, Annots: annots}
	}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{page()}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	screen := annot("Screen", " /T (sample)")
	write("plain.pdf", pdfxtest.Doc{})
	write("screen.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(screen)}})
	write("movie.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Movie", " /Movie << /F (sample.mov) >>"))}})
	write("sound-second-page.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(), page(annot("Sound", " /Sound {{obj 0}}"))}, Objects: []string{pdfxtest.Stream("<< /Type /Sound /R 8000 >>", []byte("sample"))}})
	write("rich-media-object-stream.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("RichMedia", " /RichMediaContent << >>"))}, ObjStream: true})
	write("after-link.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Link", ""), screen)}})
	write("text-note-only.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Text", " /Contents (sample)"))}})
	write("link-only.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{page(annot("Link", " /A << /S /URI /URI (https://example.org/sample.mp4) >>"))}})
	write("unlisted.pdf", pdfxtest.Doc{Objects: []string{screen}})
}
