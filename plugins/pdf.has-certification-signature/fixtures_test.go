package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const form = "<< /Fields [{{obj 0}}] /SigFlags 3 >>"

func signed(perms string) []byte {
	catalog := map[string]string{"/AcroForm": form}
	if perms != "" {
		catalog["/Perms"] = perms
	}
	return pdfxtest.Sign(pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{}},
		Catalog: catalog,
		Objects: []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
	}))
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "plain.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}))
	// The field's signature value is the certification signature.
	pdfxtest.Raw(t, "testdata", "certified.pdf", signed("<< /DocMDP {{obj 1}} >>"))
	// An approval signature: signed, but not certified.
	pdfxtest.Raw(t, "testdata", "approval.pdf", signed(""))
	// Usage rights (/UR3) alone are not a certification.
	pdfxtest.Raw(t, "testdata", "usage-rights.pdf", signed("<< /UR3 {{obj 1}} >>"))
	pdfxtest.Raw(t, "testdata", "empty-perms.pdf", signed("<< >>"))
}
