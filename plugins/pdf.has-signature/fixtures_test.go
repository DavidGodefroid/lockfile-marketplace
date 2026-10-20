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
		return map[string]string{"/AcroForm": "<< /Fields [" + fields + "] /SigFlags 3 >>"}
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("signed.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
	})
	// A document time stamp is a signature too.
	write("timestamp.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			pdfxtest.SignatureField(0),
			"<< /Type /DocTimeStamp /Filter /Adobe.PPKLite /SubFilter /ETSI.RFC3161 /ByteRange [0 1 2 3] /Contents <00> >>",
		},
	})
	// /FT /Sig set on the parent field, the value on its kid.
	write("inherited-type.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Sig /T (sample) /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}} /T (kid) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj 2}} >>",
			pdfxtest.SignatureValue(),
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
		ObjStream: true,
	})
	// A signature field waiting to be signed.
	write("empty-field.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Sig /T (sample) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
	})
	write("text-field.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Tx /T (sample) /V (sample) /DA (/Helv 0 Tf 0 g) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
	})
	// A signature dictionary no field points to.
	write("orphan-value.pdf", pdfxtest.Doc{Objects: []string{pdfxtest.SignatureValue()}})
}
