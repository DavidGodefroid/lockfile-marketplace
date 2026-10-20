package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	sheetNS      = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	relWorksheet = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"
)

// sheet returns a worksheet part in the given namespace with sheetData rows.
func sheet(ns, rows string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="` + ns + `"><sheetData>` + rows + `</sheetData></worksheet>`)
}

// workbookRels links the workbook to worksheets by target.
func workbookRels(targets ...string) fx.Entry {
	var rels [][3]string
	for _, t := range targets {
		rels = append(rels, [3]string{relWorksheet, t, ""})
	}
	return fx.File("xl/_rels/workbook.xml.rels", fx.Rels(rels...))
}

// Rows of a sheet: values only, and one with a formula cell.
const (
	values  = `<row r="1"><c r="A1"><v>1</v></c><c r="B1"><v>2</v></c></row>`
	formula = `<row r="1"><c r="A1"><v>1</v></c><c r="B1"><f>A1+1</f><v>2</v></c></row>`
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	fx.Raw(t, "testdata", "values-only.xlsx", fx.Xlsx(nil,
		workbookRels("worksheets/sheet1.xml"),
		fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, values))))
	fx.Raw(t, "testdata", "formula.xlsx", fx.Xlsx(nil,
		workbookRels("worksheets/sheet1.xml"),
		fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, formula))))
	// The formula is on the second sheet only.
	fx.Raw(t, "testdata", "second-sheet.xlsx", fx.Xlsx(nil,
		workbookRels("worksheets/sheet1.xml", "worksheets/sheet2.xml"),
		fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, values)),
		fx.File("xl/worksheets/sheet2.xml", sheet(sheetNS, formula))))
	// An absolute relationship target.
	fx.Raw(t, "testdata", "absolute-target.xlsx", fx.Xlsx(nil,
		workbookRels("/xl/worksheets/sheet1.xml"),
		fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, formula))))
	// Strict SpreadsheetML namespace and relationship type.
	fx.Raw(t, "testdata", "strict-formula.xlsx", fx.Xlsx(nil,
		fx.File("xl/_rels/workbook.xml.rels", fx.Rels([3]string{
			"http://purl.oclc.org/ooxml/officeDocument/relationships/worksheet", "worksheets/sheet1.xml", ""})),
		fx.File("xl/worksheets/sheet1.xml", sheet("http://purl.oclc.org/ooxml/spreadsheetml/main", formula))))
	// An element named f outside the SpreadsheetML namespace.
	fx.Raw(t, "testdata", "foreign-f.xlsx", fx.Xlsx(nil,
		workbookRels("worksheets/sheet1.xml"),
		fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, values+`<x:f xmlns:x="urn:example"/>`))))
	// A worksheet-shaped part with a formula the workbook does not link.
	fx.Raw(t, "testdata", "unlinked-sheet.xlsx", fx.Xlsx(nil,
		workbookRels("worksheets/sheet1.xml"),
		fx.File("xl/worksheets/sheet1.xml", sheet(sheetNS, values)),
		fx.File("xl/worksheets/sheet9.xml", sheet(sheetNS, formula))))
}
