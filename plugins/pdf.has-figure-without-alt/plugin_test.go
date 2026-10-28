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

func TestNoAnswer(t *testing.T) {
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt":         pdfxtest.Sample,
		"role-map-cycle.pdf": nested(" /RoleMap << /Pic /Image /Image /Pic >>", false, "/S /Pic"),
		"cyclic-tree.pdf": tagged(" /K {{obj 1}}", false,
			"<< /S /Sect /P {{obj 0}} /K {{obj 2}} >>",
			"<< /S /Div /P {{obj 1}} /K [{{obj 1}}] >>",
		),
		"string-type.pdf": nested("", false, "/S (Figure)"),
	})
}
