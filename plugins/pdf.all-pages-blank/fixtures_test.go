package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	for name, kinds := range map[string][]string{
		"one-blank.pdf":       {"blank"},
		"two-blank.pdf":       {"blank", "white"},
		"blank-then-text.pdf": {"blank", "text"},
		"text.pdf":            {"text"},
		"gray-fill.pdf":       {"blank", "gray"},
	} {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(pdfxtest.Pages(kinds...)))
	}
}
