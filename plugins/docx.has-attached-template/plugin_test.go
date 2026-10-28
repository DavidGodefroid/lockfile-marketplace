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
		"plain.zip":  fx.Zip(fx.File("a.txt", fx.Sample)),
		"xlsx":       fx.Xlsx(nil),
		"malformed document rels": fx.Docx(fx.Paragraph,
			fx.File("word/_rels/document.xml.rels", []byte(`<Relationships><Relationship`))),
		"malformed settings rels": fx.Docx(fx.Paragraph,
			docRels(relSettings, "settings.xml"),
			fx.File("word/settings.xml", settings),
			fx.File("word/_rels/settings.xml.rels", []byte(`<Relationships><Relationship`))),
	})
}
