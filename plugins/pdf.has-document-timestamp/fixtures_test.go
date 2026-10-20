package main

import (
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// sigValue is pdfxtest.SignatureValue with its /Type and /SubFilter
// replaced.
func sigValue(typ, subFilter string) string {
	v := pdfxtest.SignatureValue()
	v = strings.Replace(v, "/Type /Sig", "/Type "+typ, 1)
	return strings.Replace(v, "/SubFilter /adbe.pkcs7.detached", "/SubFilter "+subFilter, 1)
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
	write("signed.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
	})
	write("pades.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), sigValue("/Sig", "/ETSI.CAdES.detached")},
	})
	write("timestamp.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), sigValue("/DocTimeStamp", "/ETSI.RFC3161")},
	})
	// Either marker is enough on its own.
	write("type-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), sigValue("/DocTimeStamp", "/adbe.pkcs7.detached")},
	})
	write("subfilter-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{pdfxtest.SignatureField(0), sigValue("/Sig", "/ETSI.RFC3161")},
	})
	// A signature, then a time stamp added for long-term validation.
	write("signed-then-timestamped.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			pdfxtest.SignatureField(0), pdfxtest.SignatureValue(),
			pdfxtest.SignatureField(2), sigValue("/DocTimeStamp", "/ETSI.RFC3161"),
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{pdfxtest.SignatureField(0), sigValue("/DocTimeStamp", "/ETSI.RFC3161")},
		ObjStream: true,
	})
	// A signature field waiting for its time stamp.
	write("empty-field.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Sig /T (sample) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
	})
	// A time stamp dictionary no field points to.
	write("orphan-timestamp.pdf", pdfxtest.Doc{Objects: []string{sigValue("/DocTimeStamp", "/ETSI.RFC3161")}})
}
