package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// sigField is a signed signature field named name whose value is object at.
func sigField(name string, at int) string {
	return fmt.Sprintf("<< /FT /Sig /T (%s) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj %d}} >>", name, at)
}

// padesValue is a signature value with the PAdES sub-filter.
func padesValue() string {
	return strings.Replace(pdfxtest.SignatureValue(), "/adbe.pkcs7.detached", "/ETSI.CAdES.detached", 1)
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
	write("pades.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{sigField("sample1", 1), padesValue()},
	})
	write("adbe-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{sigField("sample1", 1), pdfxtest.SignatureValue()},
	})
	write("adbe-sha1.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			sigField("sample1", 1),
			strings.Replace(pdfxtest.SignatureValue(), "/adbe.pkcs7.detached", "/adbe.pkcs7.sha1", 1),
		},
	})
	// A document time stamp is not a PAdES signature by itself.
	write("timestamp-only.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			sigField("sample1", 1),
			"<< /Type /DocTimeStamp /Filter /Adobe.PPKLite /SubFilter /ETSI.RFC3161 /ByteRange [0 1 2 3] /Contents <00> >>",
		},
	})
	// An adbe signature followed by a PAdES one.
	write("adbe-then-pades.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}} {{obj 2}}"),
		Objects: []string{
			sigField("sample1", 1), pdfxtest.SignatureValue(),
			sigField("sample2", 3), padesValue(),
		},
	})
	// A signed kid under a parent carrying /FT /Sig.
	write("inherited-type.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{
			"<< /FT /Sig /T (sample) /Kids [{{obj 1}}] >>",
			"<< /Parent {{obj 0}} /T (kid1) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj 2}} >>",
			padesValue(),
		},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Catalog:   form("{{obj 0}}"),
		Objects:   []string{sigField("sample1", 1), padesValue()},
		ObjStream: true,
	})
	// A signature field waiting to be signed.
	write("unsigned-field.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{"<< /FT /Sig /T (sample1) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] >>"},
	})
	// A PAdES signature dictionary no field points to.
	write("orphan-pades.pdf", pdfxtest.Doc{
		Catalog: form("{{obj 0}}"),
		Objects: []string{sigField("sample1", 1), pdfxtest.SignatureValue(), padesValue()},
	})
}
