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
		"sample.docx": fx.Docx(fx.Paragraph),
		"malformed workbook rels": fx.Xlsx(nil,
			fx.File("xl/_rels/workbook.xml.rels", []byte(`<Relationships><Relationship`))),
		"malformed sheet": fx.Xlsx(nil,
			workbookRels("worksheets/sheet1.xml"),
			fx.File("xl/worksheets/sheet1.xml", []byte(`<worksheet xmlns="`+sheetNS+`"><sheetData>`))),
		// Malformed after a formula: still an error, not 1.
		"malformed sheet after formula": fx.Xlsx(nil,
			workbookRels("worksheets/sheet1.xml", "worksheets/sheet2.xml"),
			fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, formula)),
			fx.File("xl/worksheets/sheet2.xml", []byte(`<worksheet xmlns="`+sheetNS+`"><row`))),
	})
}
