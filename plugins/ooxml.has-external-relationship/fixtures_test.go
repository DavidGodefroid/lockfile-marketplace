package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// sampleURL is a synthetic external target.
const sampleURL = "https://example.invalid/sample"

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	// An external hyperlink in the main document part's rels.
	fx.Raw(t, "testdata", "external-hyperlink.docx", fx.Docx(fx.Paragraph,
		fx.File("word/_rels/document.xml.rels", fx.Rels([3]string{fx.RelHyperlink, sampleURL, "External"}))))
	// An external attached template in the settings part's rels.
	fx.Raw(t, "testdata", "external-template.docx", fx.Docx(fx.Paragraph,
		fx.File("word/_rels/settings.xml.rels", fx.Rels([3]string{fx.RelAttachedTmpl, sampleURL, "External"}))))
	// An external relationship in the package root rels.
	fx.Raw(t, "testdata", "external-root.xlsx", fx.Zip(
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/xl/workbook.xml", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"})),
		fx.File("_rels/.rels", fx.Rels(
			[3]string{fx.RelOfficeDocument, "xl/workbook.xml", ""},
			[3]string{fx.RelHyperlink, sampleURL, "External"})),
		fx.File("xl/workbook.xml", []byte(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"/>`))))
	// Internal relationships only, one with an explicit Internal mode.
	fx.Raw(t, "testdata", "internal-only.docx", fx.Docx(fx.Paragraph,
		fx.File("word/_rels/document.xml.rels", fx.Rels(
			[3]string{fx.RelComments, "comments.xml", "Internal"},
			[3]string{fx.RelHyperlink, "#sample", ""}))))
	// A TargetMode in the wrong case is not the OPC value.
	fx.Raw(t, "testdata", "lowercase-mode.docx", fx.Docx(fx.Paragraph,
		fx.File("word/_rels/document.xml.rels", fx.Rels([3]string{fx.RelHyperlink, sampleURL, "external"}))))
}
