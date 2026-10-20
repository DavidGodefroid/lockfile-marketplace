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
		"no root rels": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/xl/workbook.xml", ctWorkbook})),
			fx.File("xl/workbook.xml", workbookXML(sheetML, hidden)),
		),
		"workbook part missing": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/xl/workbook.xml", ctWorkbook})),
			fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "xl/workbook.xml", ""})),
		),
		"external workbook": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/xl/workbook.xml", ctWorkbook})),
			fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "xl/workbook.xml", "External"})),
			fx.File("xl/workbook.xml", workbookXML(sheetML, hidden)),
		),
		"binary workbook":    xlsx("application/vnd.ms-excel.sheet.binary.macroEnabled.main", workbookXML(sheetML, hidden)),
		"malformed workbook": xlsx(ctWorkbook, []byte(`<workbook xmlns="`+sheetML+`"><sheets>`)),
		"hidden then malformed": xlsx(ctWorkbook, []byte(
			`<workbook xmlns="`+sheetML+`"><sheets>`+hidden+`</sheets><definedNames`)),
		"malformed root rels": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/xl/workbook.xml", ctWorkbook})),
			fx.File("_rels/.rels", []byte(`<Relationships><Relationship`)),
			fx.File("xl/workbook.xml", workbookXML(sheetML, hidden)),
		),
	})
}
