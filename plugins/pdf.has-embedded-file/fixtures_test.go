package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

var embedded = pdfxtest.Stream("<< /Type /EmbeddedFile >>", []byte("sample"))

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("names-tree.pdf", pdfxtest.Doc{
		Catalog: map[string]string{"/Names": "<< /EmbeddedFiles << /Names [(sample.txt) {{obj 0}}] >> >>"},
		Objects: []string{"<< /Type /Filespec /F (sample.txt) /UF (sample.txt) /EF << /F {{obj 1}} >> >>", embedded},
	})
	write("file-attachment.pdf", pdfxtest.Doc{
		Pages: []pdfxtest.Page{{Annots: []string{
			"<< /Type /Annot /Subtype /FileAttachment /Rect [0 0 10 10] /FS {{obj 0}} >>",
		}}},
		Objects: []string{"<< /Type /Filespec /F (sample.txt) /EF << /F {{obj 1}} >> >>", embedded},
	})
	write("orphan-stream.pdf", pdfxtest.Doc{Objects: []string{embedded}})
	write("object-stream.pdf", pdfxtest.Doc{
		Objects:   []string{"<< /Type /Filespec /F (sample.txt) /EF << /F {{obj 1}} >> >>", embedded},
		ObjStream: true,
	})
	// A link to another file: a file specification without /EF.
	write("external-file.pdf", pdfxtest.Doc{Pages: []pdfxtest.Page{{Annots: []string{
		"<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] /A << /S /GoToR /F << /Type /Filespec /F (sample.pdf) >> /D [0 /Fit] >> >>",
	}}}})
}
