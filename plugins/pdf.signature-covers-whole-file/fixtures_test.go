package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	signed := func(d pdfxtest.Doc) []byte {
		d.Pages = page
		return pdfxtest.Sign(pdfxtest.Write(d))
	}
	form := map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] /SigFlags 3 >>"}
	sig := []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()}
	update := func(b []byte) []byte {
		return pdfxtest.AppendUpdate(b, []string{"<< /Producer (sample) >>"}, "/Info {{obj 0}}")
	}
	pdfxtest.Raw(t, "testdata", "plain.pdf", signed(pdfxtest.Doc{}))
	pdfxtest.Raw(t, "testdata", "signed.pdf", signed(pdfxtest.Doc{Catalog: form, Objects: sig}))
	pdfxtest.Raw(t, "testdata", "object-stream.pdf", signed(pdfxtest.Doc{Catalog: form, Objects: sig, ObjStream: true}))
	// Changed after signing: the signature covers the first revision only.
	pdfxtest.Raw(t, "testdata", "updated-after-signing.pdf", update(signed(pdfxtest.Doc{Catalog: form, Objects: sig})))
	pdfxtest.Raw(t, "testdata", "object-stream-updated.pdf", update(signed(pdfxtest.Doc{Catalog: form, Objects: sig, ObjStream: true})))
	// A byte range that does not reach the end of the file.
	pdfxtest.Raw(t, "testdata", "partial-range.pdf", signed(pdfxtest.Doc{
		Catalog: form,
		Objects: []string{
			pdfxtest.SignatureField(0),
			"<< /Type /Sig /Filter /Adobe.PPKLite /SubFilter /adbe.pkcs7.detached /ByteRange [0 1 2 3] /Contents <00> >>",
		},
	}))
	// A signature field waiting to be signed.
	pdfxtest.Raw(t, "testdata", "empty-field.pdf", signed(pdfxtest.Doc{
		Catalog: form,
		Objects: []string{"<< /FT /Sig /T (sample) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
	}))
}
