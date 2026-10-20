package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// run wraps run properties and sample text in a paragraph of a w:body.
func run(rPr string) string {
	return `<w:body><w:p><w:r><w:rPr>` + rPr + `</w:rPr><w:t>sample</w:t></w:r></w:p></w:body>`
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "vanish.docx", fx.Docx(run(`<w:vanish/>`)))
	fx.Raw(t, "testdata", "vanish-true.docx", fx.Docx(run(`<w:vanish w:val="true"/>`)))
	// A paragraph mark made hidden, in the paragraph's run properties.
	fx.Raw(t, "testdata", "paragraph-mark.docx", fx.Docx(
		`<w:body><w:p><w:pPr><w:rPr><w:vanish/></w:rPr></w:pPr><w:r><w:t>sample</w:t></w:r></w:p></w:body>`))
	// A vanish explicitly turned off is not hidden text.
	fx.Raw(t, "testdata", "vanish-off.docx", fx.Docx(run(`<w:vanish w:val="0"/>`)))
	fx.Raw(t, "testdata", "vanish-false.docx", fx.Docx(run(`<w:vanish w:val="false"/>`)))
	// Other run properties only.
	fx.Raw(t, "testdata", "bold.docx", fx.Docx(run(`<w:b/>`)))
	// A vanish in the strict WordprocessingML namespace.
	fx.Raw(t, "testdata", "strict-vanish.docx", fx.Zip(
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
		fx.File("word/document.xml", []byte(`<w:document xmlns:w="http://purl.oclc.org/ooxml/wordprocessingml/main">`+run(`<w:vanish/>`)+`</w:document>`)),
	))
	// An element named vanish outside the WordprocessingML namespace.
	fx.Raw(t, "testdata", "foreign-vanish.docx", fx.Docx(
		`<w:body><w:p><x:vanish xmlns:x="urn:example"/></w:p></w:body>`))
	// A hidden style in word/styles.xml: that part is not read.
	fx.Raw(t, "testdata", "styles-vanish.docx", fx.Docx(fx.Paragraph,
		fx.File("word/styles.xml", fx.WordXML("styles", `<w:style w:styleId="s"><w:rPr><w:vanish/></w:rPr></w:style>`))))
}
