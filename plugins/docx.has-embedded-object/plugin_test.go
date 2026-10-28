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
		"no root rels": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
			fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph))),
		"missing main part": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
			fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "word/document.xml", ""}))),
		"malformed root rels": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
			fx.File("_rels/.rels", []byte(`<Relationships`)),
			fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph))),
		"malformed document rels": fx.Docx(fx.Paragraph,
			fx.File("word/_rels/document.xml.rels", []byte(`<Relationships><Relationship`))),
	})
}
