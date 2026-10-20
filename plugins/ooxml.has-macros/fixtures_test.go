package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// Content types of macro parts.
const (
	ctVBAProject = "application/vnd.ms-office.vbaProject"
	ctMacrosheet = "application/vnd.ms-excel.macrosheet+xml"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	// A VBA project found by its name alone, with no content type.
	fx.Raw(t, "testdata", "vba-name.docm", fx.Docx(fx.Paragraph,
		fx.File("word/vbaProject.bin", fx.Sample)))
	// A VBA project found by its override content type, matched
	// case-insensitively, under another name.
	fx.Raw(t, "testdata", "vba-override.xlsm", fx.Xlsx(
		[][2]string{{"/xl/macros.bin", "APPLICATION/VND.MS-OFFICE.VBAPROJECT"}},
		fx.File("xl/macros.bin", fx.Sample)))
	// A VBA project found by a Default content type for its extension.
	fx.Raw(t, "testdata", "vba-default.docm", fx.Zip(
		fx.File("[Content_Types].xml", []byte(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`+
			`<Default Extension="vba" ContentType="`+ctVBAProject+`"/>`+
			`<Override PartName="/word/document.xml" ContentType="`+fx.DocxMain+`"/></Types>`)),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "word/document.xml", ""})),
		fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph)),
		fx.File("word/code.vba", fx.Sample)))
	// An Excel 4.0 macro sheet.
	fx.Raw(t, "testdata", "macrosheet.xlsm", fx.Xlsx(
		[][2]string{{"/xl/macrosheets/sheet1.xml", ctMacrosheet}},
		fx.File("xl/macrosheets/sheet1.xml", []byte(`<xm:macrosheet xmlns:xm="http://schemas.microsoft.com/office/excel/2006/main"/>`))))
	// A Default content type for VBA, as macro-enabled templates declare
	// it, with no part using that extension.
	fx.Raw(t, "testdata", "unused-default.docx", fx.Zip(
		fx.File("[Content_Types].xml", []byte(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`+
			`<Default Extension="bin" ContentType="`+ctVBAProject+`"/>`+
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`+
			`<Override PartName="/word/document.xml" ContentType="`+fx.DocxMain+`"/></Types>`)),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "word/document.xml", ""})),
		fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph))))
	// An override naming a macro sheet that is not in the package.
	fx.Raw(t, "testdata", "missing-override.xlsx", fx.Xlsx(
		[][2]string{{"/xl/macrosheets/sheet1.xml", ctMacrosheet}}))
}
