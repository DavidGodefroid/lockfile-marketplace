package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	pdfxtest.Raw(t, "testdata", "plain.pdf", plain)
	pdfxtest.Raw(t, "testdata", "all-allowed.pdf", pdfxtest.EncryptPerms(t, plain, "", "sample-owner", pdfxtest.PermAll))
	pdfxtest.Raw(t, "testdata", "printing-allowed.pdf", pdfxtest.EncryptPerms(t, plain, "", "sample-owner", pdfxtest.PermPrint))
	pdfxtest.Raw(t, "testdata", "printing-denied.pdf", pdfxtest.EncryptPerms(t, plain, "", "sample-owner", pdfxtest.PermCopy))
	pdfxtest.Raw(t, "testdata", "none-allowed.pdf", pdfxtest.EncryptPerms(t, plain, "", "sample-owner", pdfxtest.PermNone))
}
