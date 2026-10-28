package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// Content types of workbook parts.
const (
	ctWorkbook      = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	ctWorkbookMacro = "application/vnd.ms-excel.sheet.macroEnabled.main+xml"
)

// workbookXML is a workbook part in namespace ns whose sheets element
// holds sheets.
func workbookXML(ns, sheets string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<workbook xmlns="` + ns + `" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<sheets>` + sheets + `</sheets></workbook>`)
}

// xlsx returns a package whose workbook part, of content type ct, is
// workbook, plus extra entries.
func xlsx(ct string, workbook []byte, extra ...fx.Entry) []byte {
	entries := []fx.Entry{
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/xl/workbook.xml", ct})),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "xl/workbook.xml", ""})),
		fx.File("xl/workbook.xml", workbook),
	}
	return fx.Zip(append(entries, extra...)...)
}

const (
	visible    = `<sheet name="Sheet1" sheetId="1" r:id="rId1"/>`
	explicit   = `<sheet name="Sheet2" sheetId="2" state="visible" r:id="rId2"/>`
	hidden     = `<sheet name="Sheet3" sheetId="3" state="hidden" r:id="rId3"/>`
	veryHidden = `<sheet name="Sheet4" sheetId="4" state="veryHidden" r:id="rId4"/>`
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty-sheets.xlsx", fx.Xlsx(nil))
	fx.Raw(t, "testdata", "visible.xlsx", xlsx(ctWorkbook, workbookXML(sheetML, visible+explicit)))
	fx.Raw(t, "testdata", "hidden.xlsx", xlsx(ctWorkbook, workbookXML(sheetML, visible+hidden)))
	fx.Raw(t, "testdata", "very-hidden.xlsm", xlsx(ctWorkbookMacro, workbookXML(sheetML, veryHidden+visible)))
	fx.Raw(t, "testdata", "strict-hidden.xlsx", xlsx(ctWorkbook, workbookXML(sheetMLStrict, hidden)))
	// A state attribute on an element that is not a sheet does not count.
	fx.Raw(t, "testdata", "state-elsewhere.xlsx", xlsx(ctWorkbook, []byte(
		`<workbook xmlns="`+sheetML+`"><bookViews><workbookView state="hidden"/></bookViews>`+
			`<sheets>`+visible+`</sheets></workbook>`)))
	// A sheet element in a foreign namespace does not count.
	fx.Raw(t, "testdata", "foreign-sheet.xlsx", xlsx(ctWorkbook, []byte(
		`<workbook xmlns="`+sheetML+`" xmlns:x="urn:sample"><sheets>`+visible+`</sheets>`+
			`<extLst><ext><x:sheet state="hidden"/></ext></extLst></workbook>`)))
	// A hidden sheet in a worksheet-like part outside the workbook part
	// does not count.
	fx.Raw(t, "testdata", "hidden-outside-workbook.xlsx", xlsx(ctWorkbook, workbookXML(sheetML, visible),
		fx.File("xl/other.xml", workbookXML(sheetML, hidden))))
}
