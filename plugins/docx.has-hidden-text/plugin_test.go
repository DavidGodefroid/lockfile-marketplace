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
		"empty.bin":   nil,
		"sample.txt":  fx.Sample,
		"plain.zip":   fx.Zip(fx.File("a.txt", fx.Sample)),
		"sample.xlsx": fx.Xlsx(nil),
		"malformed document": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
			fx.File("word/document.xml", []byte(`<w:document xmlns:w="`+fx.WNS+`"><w:body>`)),
		),
		// Malformed after a vanish: still an error, not 1.
		"malformed after vanish": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
			fx.File("word/document.xml", []byte(`<w:document xmlns:w="`+fx.WNS+`"><w:body><w:vanish/><w:p>`)),
		),
		"no main document": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("word/styles.xml", fx.WordXML("styles", `<w:vanish/>`)),
		),
	})
}
