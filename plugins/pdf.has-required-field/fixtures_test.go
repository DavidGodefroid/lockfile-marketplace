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
	form := func(fields string) map[string]string {
		return map[string]string{"/AcroForm": "<< /Fields [" + fields + "] /DA (/Helv 0 Tf 0 g) >>"}
	}
	widget := " /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /DA (/Helv 0 Tf 0 g)"
	write("plain.pdf", pdfxtest.Doc{})
	write("required.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 2" + widget + " >>"},
	})
	// Multiline (bit 13) and Required (bit 2).
	write("required-among-flags.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 4098" + widget + " >>"},
	})
	write("second-field-required.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 1}}"),
		Objects: []string{
			"<< /FT /Tx /T (sample) /Ff 0" + widget + " >>",
			"<< /FT /Btn /T (sample2) /V /Off /Ff 2" + widget + " >>",
		},
	})
	// Required set on the parent field only, the widget is its kid.
	write("required-parent.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Tx /T (sample) /Ff 2 /DA (/Helv 0 Tf 0 g) /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}}" + widget + " >>",
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{"<< /FT /Tx /T (sample) /Ff 2" + widget + " >>"},
		ObjStream: true,
	})
	write("read-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 1" + widget + " >>"},
	})
	write("no-flags.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /V (sample)" + widget + " >>"},
	})
	write("empty-form.pdf", pdfxtest.Doc{Catalog: form("")})
	// A required widget annotation no field tree points to.
	write("orphan-field.pdf", pdfxtest.Doc{
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 2" + widget + " >>"},
	})
}
