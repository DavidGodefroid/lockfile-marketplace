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
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Sign(pdfxtest.Write(d)))
	}
	form := func(fields string) map[string]string {
		return map[string]string{"/AcroForm": "<< /Fields [" + fields + "] /SigFlags 1 >>"}
	}
	widget := " /Type /Annot /Subtype /Widget /Rect [0 0 0 0]"
	write("plain.pdf", pdfxtest.Doc{})
	write("unsigned.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Sig /T (sample)" + widget + " >>"},
	})
	write("null-value.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Sig /T (sample) /V null" + widget + " >>"},
	})
	write("signed.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
	})
	// One signed field, one empty field.
	write("signed-and-unsigned.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			pdfxtest.SignatureField(0),
			pdfxtest.SignatureValue(),
			"<< /FT /Sig /T (sample2)" + widget + " >>",
		},
	})
	// /FT /Sig set on the parent field only, the widget is its kid.
	write("inherited-type.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Sig /T (sample) /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}}" + widget + " >>",
		},
	})
	// The parent field carries the value, its widget kid does not.
	write("inherited-value.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Sig /T (sample) /V {{obj 2}} /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}}" + widget + " >>",
			pdfxtest.SignatureValue(),
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{"<< /FT /Sig /T (sample)" + widget + " >>"},
		ObjStream: true,
	})
	// An empty text field is not a signature field.
	write("text-only.pdf", pdfxtest.Doc{
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] /DA (/Helv 0 Tf 0 g) >>"},
		Objects: []string{"<< /FT /Tx /T (sample)" + widget + " /DA (/Helv 0 Tf 0 g) >>"},
	})
	write("empty-form.pdf", pdfxtest.Doc{Catalog: form("")})
	// A signature widget no field tree points to.
	write("orphan-field.pdf", pdfxtest.Doc{
		Objects: []string{"<< /FT /Sig /T (sample)" + widget + " >>"},
	})
}
