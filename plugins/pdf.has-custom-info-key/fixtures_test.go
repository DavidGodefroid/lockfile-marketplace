package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func doc(info map[string]string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Info: info})
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-info.pdf", doc(nil))
	pdfxtest.Raw(t, "testdata", "empty-info.pdf", doc(map[string]string{}))
	pdfxtest.Raw(t, "testdata", "standard-keys.pdf", doc(map[string]string{
		"/Title": "(sample)", "/Author": "(sample)", "/Subject": "(sample)", "/Keywords": "(sample)",
		"/Creator": "(sample)", "/Producer": "(sample)", "/CreationDate": "(D:20260101000000Z)",
		"/ModDate": "(D:20260101000000Z)", "/Trapped": "/False",
	}))
	pdfxtest.Raw(t, "testdata", "custom-key.pdf", doc(map[string]string{"/Title": "(sample)", "/SampleKey": "(sample)"}))
	pdfxtest.Raw(t, "testdata", "custom-key-empty-value.pdf", doc(map[string]string{"/SampleKey": "()"}))
	pdfxtest.Raw(t, "testdata", "custom-key-null.pdf", doc(map[string]string{"/Producer": "(sample)", "/SampleKey": "null"}))
	pdfxtest.Raw(t, "testdata", "lowercase-standard-name.pdf", doc(map[string]string{"/title": "(sample)"}))
}
