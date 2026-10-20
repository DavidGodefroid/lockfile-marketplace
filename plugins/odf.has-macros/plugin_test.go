package main

import (
	"os"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
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
		"empty.bin":  nil,
		"sample.txt": fx.Sample,
		"plain.zip":  fx.Zip(fx.File("Basic/Standard/Module1.xml", fx.Sample)),
		"mimetype not first": fx.Zip(
			fx.File("content.xml", fx.Sample),
			mimetype("application/vnd.oasis.opendocument.text"),
			fx.File("Basic/Standard/Module1.xml", fx.Sample)),
		"epub": fx.Zip(
			mimetype("application/epub+zip"),
			fx.File("Scripts/sample.js", fx.Sample)),
		"bare prefix": fx.Zip(
			mimetype("application/vnd.oasis.opendocument."),
			fx.File("Basic/Standard/Module1.xml", fx.Sample)),
		"docx": fx.Docx(fx.Paragraph),
	})
}
