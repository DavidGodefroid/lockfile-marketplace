package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// A comment anchored in the body, referencing comment 0.
const anchored = `<w:body><w:p><w:commentRangeStart w:id="0"/><w:r><w:t>sample</w:t></w:r>` +
	`<w:commentRangeEnd w:id="0"/><w:r><w:commentReference w:id="0"/></w:r></w:p></w:body>`

// One comment of a comments part.
const comment = `<w:comment w:id="0" w:author="A"><w:p><w:r><w:t>note</w:t></w:r></w:p></w:comment>`

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "comments.docx", fx.Docx(anchored,
		fx.File("word/comments.xml", fx.WordXML("comments", comment))))
	// The comments part under another name, upper case.
	fx.Raw(t, "testdata", "renamed.docx", fx.Docx(anchored,
		fx.File("WORD/Notes.XML", fx.WordXML("comments", comment))))
	// The strict WordprocessingML namespace.
	fx.Raw(t, "testdata", "strict.docx", fx.Docx(fx.Paragraph,
		fx.File("word/comments.xml", []byte(`<w:comments xmlns:w="http://purl.oclc.org/ooxml/wordprocessingml/main">`+comment+`</w:comments>`))))
	// An empty comments root.
	fx.Raw(t, "testdata", "empty-comments.docx", fx.Docx(fx.Paragraph,
		fx.File("word/comments.xml", fx.WordXML("comments", ""))))
	// References in the body with no comments part.
	fx.Raw(t, "testdata", "references-only.docx", fx.Docx(anchored))
	// A comment element in another namespace.
	fx.Raw(t, "testdata", "other-namespace.docx", fx.Docx(fx.Paragraph,
		fx.File("word/custom.xml", []byte(`<x:root xmlns:x="urn:example"><x:comment/></x:root>`))))
	// A comments part outside word/*.xml.
	fx.Raw(t, "testdata", "nested.docx", fx.Docx(fx.Paragraph,
		fx.File("word/glossary/comments.xml", fx.WordXML("comments", comment)),
		fx.File("customXml/comments.xml", fx.WordXML("comments", comment))))
}
