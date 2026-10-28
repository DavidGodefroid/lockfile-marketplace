package main

import (
	"fmt"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// sigField is a signed signature field named name whose value is object at.
func sigField(name string, at int) string {
	return fmt.Sprintf("<< /FT /Sig /T (%s) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj %d}} >>", name, at)
}

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
	write("one-signature.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{sigField("sample1", 1), pdfxtest.SignatureValue()},
	})
	write("two-signatures.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			sigField("sample1", 1), pdfxtest.SignatureValue(),
			sigField("sample2", 3), pdfxtest.SignatureValue(),
		},
	})
	write("three-signatures.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}} {{obj 4}}"),
		Objects: []string{
			sigField("sample1", 1), pdfxtest.SignatureValue(),
			sigField("sample2", 3), pdfxtest.SignatureValue(),
			sigField("sample3", 5), pdfxtest.SignatureValue(),
		},
	})
	// A document time stamp counts as a second signature.
	write("signature-and-timestamp.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			sigField("sample1", 1), pdfxtest.SignatureValue(),
			sigField("sample2", 3),
			"<< /Type /DocTimeStamp /Filter /Adobe.PPKLite /SubFilter /ETSI.RFC3161 /ByteRange [0 1 2 3] /Contents <00> >>",
		},
	})
	// Two signed kids under one parent carrying /FT /Sig.
	write("inherited-type.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Sig /T (sample) /Kids [{{obj 1}} {{obj 3}}] >>",
			"<< /Parent {{obj 0}} /T (kid1) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj 2}} >>",
			pdfxtest.SignatureValue(),
			"<< /Parent {{obj 0}} /T (kid2) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj 4}} >>",
			pdfxtest.SignatureValue(),
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			sigField("sample1", 1), pdfxtest.SignatureValue(),
			sigField("sample2", 3), pdfxtest.SignatureValue(),
		},
		ObjStream: true,
	})
	// One signed field and one waiting to be signed.
	write("one-signed-one-empty.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			sigField("sample1", 1), pdfxtest.SignatureValue(),
			"<< /FT /Sig /T (sample2) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>",
		},
	})
	// One signed field and a second signature dictionary no field points to.
	write("one-signed-one-orphan.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{sigField("sample1", 1), pdfxtest.SignatureValue(), pdfxtest.SignatureValue()},
	})
}
