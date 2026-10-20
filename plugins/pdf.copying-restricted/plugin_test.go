package main

import (
	"os"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk/harness"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestConformance(t *testing.T) {
	writeFixtures(t)
	harness.Check(t, ".")
}

// TestNoAnswer: a PDF that needs a user password has no readable
// permissions, so it gets no answer either.
func TestNoAnswer(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt":        pdfxtest.Sample,
		"empty":             nil,
		"user-password.pdf": pdfxtest.EncryptPerms(t, plain, "sample-user", "sample-owner", pdfxtest.PermNone),
	})
}
