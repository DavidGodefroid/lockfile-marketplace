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
	// Change tracking turned on, with no revision recorded yet.
	fx.Raw(t, "testdata", "track-revisions-only.docx", fx.Docx(fx.Paragraph,
		fx.File("word/settings.xml", fx.WordXML("settings", `<w:trackRevisions/>`))))
	fx.Raw(t, "testdata", "insertion.docx", fx.Docx(body(
		`<w:ins w:id="1" w:author="a"><w:r><w:t>new</w:t></w:r></w:ins>`)))
	fx.Raw(t, "testdata", "deletion.docx", fx.Docx(body(
		`<w:del w:id="1" w:author="a"><w:r><w:delText>old</w:delText></w:r></w:del>`)))
	fx.Raw(t, "testdata", "move.docx", fx.Docx(body(
		`<w:moveFrom w:id="1" w:author="a"><w:r><w:t>moved</w:t></w:r></w:moveFrom>`)))
	fx.Raw(t, "testdata", "cell-deletion.docx", fx.Docx(
		`<w:body><w:tbl><w:tr><w:tc><w:tcPr><w:cellDel w:id="1" w:author="a"/></w:tcPr><w:p/></w:tc></w:tr></w:tbl></w:body>`))
	// A formatting change: a name ending in "Change".
	fx.Raw(t, "testdata", "format-change.docx", fx.Docx(body(
		`<w:r><w:rPr><w:b/><w:rPrChange w:id="1" w:author="a"><w:rPr/></w:rPrChange></w:rPr><w:t>bold</w:t></w:r>`)))
	// A revision in another word/*.xml part than the main document.
	fx.Raw(t, "testdata", "header-insertion.docx", fx.Docx(fx.Paragraph,
		fx.File("word/header1.xml", fx.WordXML("hdr", `<w:p><w:ins w:id="1" w:author="a"/></w:p>`))))
	// A revision in the strict WordprocessingML namespace.
	fx.Raw(t, "testdata", "strict-insertion.docx", fx.Docx(fx.Paragraph,
		fx.File("word/footnotes.xml", []byte(`<w:footnotes xmlns:w="http://purl.oclc.org/ooxml/wordprocessingml/main"><w:ins/></w:footnotes>`))))
	// An element named ins outside the WordprocessingML namespace.
	fx.Raw(t, "testdata", "foreign-ins.docx", fx.Docx(fx.Paragraph,
		fx.File("word/custom.xml", []byte(`<x:root xmlns:x="urn:example"><x:ins/><x:formatChange/></x:root>`))))
	// A revision in a part outside word/*.xml.
	fx.Raw(t, "testdata", "outside-word.docx", fx.Docx(fx.Paragraph,
		fx.File("customXml/item1.xml", fx.WordXML("document", body(`<w:ins/>`)))))
}
