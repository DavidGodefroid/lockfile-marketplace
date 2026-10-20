package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// body wraps paragraph content in a w:body.
func body(p string) string {
	return `<w:body><w:p>` + p + `</w:p></w:body>`
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "simple-field.docx", fx.Docx(body(
		`<w:fldSimple w:instr="PAGE"><w:r><w:t>1</w:t></w:r></w:fldSimple>`)))
	fx.Raw(t, "testdata", "complex-field.docx", fx.Docx(body(
		`<w:r><w:fldChar w:fldCharType="begin"/></w:r>`+
			`<w:r><w:instrText xml:space="preserve"> PAGE </w:instrText></w:r>`+
			`<w:r><w:fldChar w:fldCharType="separate"/></w:r><w:r><w:t>1</w:t></w:r>`+
			`<w:r><w:fldChar w:fldCharType="end"/></w:r>`)))
	// A field in the strict WordprocessingML namespace.
	fx.Raw(t, "testdata", "strict-field.docx", fx.Zip(
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "word/document.xml", ""})),
		fx.File("word/document.xml", []byte(`<w:document xmlns:w="http://purl.oclc.org/ooxml/wordprocessingml/main"><w:body><w:p><w:fldSimple/></w:p></w:body></w:document>`))))
	// A field character with no instruction element.
	fx.Raw(t, "testdata", "fldchar-only.docx", fx.Docx(body(
		`<w:r><w:fldChar w:fldCharType="begin"/></w:r><w:r><w:fldChar w:fldCharType="end"/></w:r>`)))
	// Elements named like fields outside the WordprocessingML namespace.
	fx.Raw(t, "testdata", "foreign-field.docx", fx.Docx(body(
		`<x:fldSimple xmlns:x="urn:example"/><x:instrText xmlns:x="urn:example"/>`)))
	// A field in a header part only: not the main document.
	fx.Raw(t, "testdata", "header-field.docx", fx.Docx(fx.Paragraph,
		fx.File("word/header1.xml", fx.WordXML("hdr", `<w:p><w:fldSimple w:instr="PAGE"/></w:p>`))))
}
