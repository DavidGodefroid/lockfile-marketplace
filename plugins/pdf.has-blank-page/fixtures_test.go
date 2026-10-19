package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	for name, kinds := range map[string][]string{
		"blank.pdf":                {"blank"},
		"text.pdf":                 {"text"},
		"text-then-blank.pdf":      {"text", "blank"},
		"white-fill.pdf":           {"white"},
		"white-fill-then-text.pdf": {"white-then-text"},
		"gray-fill.pdf":            {"gray"},
		"image.pdf":                {"image"},
		"invisible-text.pdf":       {"text", "invisible-text"},
	} {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(pdfxtest.Pages(kinds...)))
	}
}
