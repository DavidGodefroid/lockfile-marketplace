package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(catalog map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: catalog})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-dss.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "dss.pdf", doc(map[string]string{"/DSS": "<< /Certs [] /OCSPs [] /CRLs [] >>"}))
	pdfxtest.Raw(t, "testdata", "empty-dss.pdf", doc(map[string]string{"/DSS": "<< >>"}))
	pdfxtest.Raw(t, "testdata", "null-dss.pdf", doc(map[string]string{"/DSS": "null"}))
}
