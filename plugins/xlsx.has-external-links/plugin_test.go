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
		"docx":       fx.Docx(fx.Paragraph),
		"no root rels": fx.Zip(fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("xl/workbook.xml", workbookXML)),
		"dangling workbook": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "xl/workbook.xml", ""}))),
		"malformed root rels": fx.Zip(fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("_rels/.rels", []byte(`<Relationships><Relationship`))),
		"malformed workbook rels": fx.Xlsx(nil,
			fx.File("xl/_rels/workbook.xml.rels", []byte(`<Relationships><Relationship`))),
	})
}
