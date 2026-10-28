package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	relBase        = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/"
	relExternal    = relBase + "externalLink"
	relWorksheet   = relBase + "worksheet"
	strictExternal = "http://purl.oclc.org/ooxml/officeDocument/relationships/externalLink"
	xlsmMain       = "application/vnd.ms-excel.sheet.macroEnabled.main+xml"
)

// workbookXML is a workbook part with no sheet.
var workbookXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
	`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheets/></workbook>`)

// externalPart is an external link part with no cached data.
var externalPart = fx.File("xl/externalLinks/externalLink1.xml", []byte(`<?xml version="1.0" encoding="UTF-8"?>`+
	`<externalLink xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"/>`))

// wbRels returns the .rels part name with one relationship of type typ.
func wbRels(name, typ, target string) fx.Entry {
	return fx.File(name, fx.Rels([3]string{typ, target, ""}))
}

// pkg returns a spreadsheet whose root relationship targets target, its
// workbook written at name with content type ct, plus extra entries.
func pkg(target, name, ct string, extra ...fx.Entry) []byte {
	entries := []fx.Entry{
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/" + name, ct})),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, target, ""})),
		fx.File(name, workbookXML),
	}
	return fx.Zip(append(entries, extra...)...)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	fx.Raw(t, "testdata", "linked.xlsx", fx.Xlsx(nil,
		wbRels("xl/_rels/workbook.xml.rels", relExternal, "externalLinks/externalLink1.xml"),
		externalPart))
	fx.Raw(t, "testdata", "strict.xlsx", fx.Xlsx(nil,
		wbRels("xl/_rels/workbook.xml.rels", strictExternal, "externalLinks/externalLink1.xml"),
		externalPart))
	// A macro-enabled workbook.
	fx.Raw(t, "testdata", "macro.xlsm", pkg("xl/workbook.xml", "xl/workbook.xml", xlsmMain,
		wbRels("xl/_rels/workbook.xml.rels", relExternal, "externalLinks/externalLink1.xml"),
		externalPart))
	// The workbook part under another name and case, with an absolute target.
	fx.Raw(t, "testdata", "renamed.xlsx", pkg("/book/Main.xml", "BOOK/MAIN.XML",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml",
		wbRels("book/_rels/main.xml.rels", relExternal, "links/link1.xml")))
	// Workbook relationships, but none to an external link.
	fx.Raw(t, "testdata", "worksheet-only.xlsx", fx.Xlsx(nil,
		wbRels("xl/_rels/workbook.xml.rels", relWorksheet, "worksheets/sheet1.xml")))
	// An external link relationship held by a worksheet, not the workbook.
	fx.Raw(t, "testdata", "on-sheet.xlsx", fx.Xlsx(nil,
		wbRels("xl/worksheets/_rels/sheet1.xml.rels", relExternal, "../externalLinks/externalLink1.xml"),
		externalPart))
}
