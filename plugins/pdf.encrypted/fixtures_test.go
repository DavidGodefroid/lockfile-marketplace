package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	pdfxtest.Raw(t, "testdata", "plain.pdf", plain)
	pdfxtest.Raw(t, "testdata", "user-password.pdf", pdfxtest.Encrypt(t, plain, "sample-user", "sample-owner"))
	pdfxtest.Raw(t, "testdata", "owner-only.pdf", pdfxtest.Encrypt(t, plain, "", "sample-owner"))
}
