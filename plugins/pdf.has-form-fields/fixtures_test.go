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
	write("text.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /V (sample)" + widget + " >>"},
	})
	write("button.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Btn /T (sample) /V /Off" + widget + " >>"},
	})
	write("choice.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Ch /T (sample) /Opt [(a) (b)]" + widget + " >>"},
	})
	// /FT /Tx set on the parent field only, the widget is its kid.
	write("inherited-type.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Tx /T (sample) /DA (/Helv 0 Tf 0 g) /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}}" + widget + " >>",
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{"<< /FT /Tx /T (sample)" + widget + " >>"},
		ObjStream: true,
	})
	// A signed signature field is not a fillable field.
	write("signature-only.pdf", pdfxtest.Doc{
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] /SigFlags 3 >>"},
		Objects: []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
	})
	write("empty-form.pdf", pdfxtest.Doc{Catalog: form("")})
	// A widget annotation no field tree points to.
	write("orphan-field.pdf", pdfxtest.Doc{
		Objects: []string{"<< /FT /Tx /T (sample)" + widget + " >>"},
	})
}
