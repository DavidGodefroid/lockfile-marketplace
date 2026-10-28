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
	write("read-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 1" + widget + " >>"},
	})
	// ReadOnly combined with Required (bit 2) and Multiline (bit 13).
	write("read-only-mixed-flags.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 4099" + widget + " >>"},
	})
	// The flag is set on the parent field, the widget is its kid.
	write("read-only-parent.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Tx /T (sample) /Ff 1 /DA (/Helv 0 Tf 0 g) /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}}" + widget + " >>",
		},
	})
	write("read-only-indirect-flags.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff {{obj 1}}" + widget + " >>", "1"},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{"<< /FT /Btn /T (sample) /Ff 1 /V /Off" + widget + " >>"},
		ObjStream: true,
	})
	write("editable.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample)" + widget + " >>"},
	})
	// Required (bit 2) only.
	write("required-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 2" + widget + " >>"},
	})
	write("empty-form.pdf", pdfxtest.Doc{Catalog: form("")})
	// A read-only widget annotation no field tree points to.
	write("orphan-field.pdf", pdfxtest.Doc{
		Objects: []string{"<< /FT /Tx /T (sample) /Ff 1" + widget + " >>"},
	})
}
