package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	relSettings       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	relSettingsStrict = "http://purl.oclc.org/ooxml/officeDocument/relationships/settings"
)

// A read-only protection setting.
const protection = `<w:documentProtection w:edit="readOnly" w:enforcement="1"/>`

// documentRels returns the relationships part of word/document.xml with one
// settings relationship to target.
func documentRels(target string) fx.Entry {
	return fx.File("word/_rels/document.xml.rels", fx.Rels([3]string{relSettings, target, ""}))
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "protected.docx", fx.Docx(fx.Paragraph,
		documentRels("settings.xml"),
		fx.File("word/settings.xml", fx.WordXML("settings", protection))))
	// Protection declared but not enforced is still declared.
	fx.Raw(t, "testdata", "not-enforced.docx", fx.Docx(fx.Paragraph,
		documentRels("settings.xml"),
		fx.File("word/settings.xml", fx.WordXML("settings", `<w:documentProtection w:edit="forms"/>`))))
	// The strict namespace and relationship type.
	fx.Raw(t, "testdata", "strict.docx", fx.Docx(fx.Paragraph,
		fx.File("word/_rels/document.xml.rels", fx.Rels([3]string{relSettingsStrict, "settings.xml", ""})),
		fx.File("word/settings.xml", []byte(`<w:settings xmlns:w="http://purl.oclc.org/ooxml/wordprocessingml/main">`+protection+`</w:settings>`))))
	// The settings part under another name, in a sub-folder.
	fx.Raw(t, "testdata", "renamed.docx", fx.Docx(fx.Paragraph,
		documentRels("config/options.xml"),
		fx.File("word/config/options.xml", fx.WordXML("settings", protection))))
	// An absolute relationship target.
	fx.Raw(t, "testdata", "absolute-target.docx", fx.Docx(fx.Paragraph,
		documentRels("/word/settings.xml"),
		fx.File("word/settings.xml", fx.WordXML("settings", protection))))
	fx.Raw(t, "testdata", "unprotected.docx", fx.Docx(fx.Paragraph,
		documentRels("settings.xml"),
		fx.File("word/settings.xml", fx.WordXML("settings", `<w:zoom w:percent="100"/>`))))
	// A settings part that the main document does not reference.
	fx.Raw(t, "testdata", "unreferenced.docx", fx.Docx(fx.Paragraph,
		fx.File("word/settings.xml", fx.WordXML("settings", protection))))
	// A settings relationship to a missing part.
	fx.Raw(t, "testdata", "dangling.docx", fx.Docx(fx.Paragraph,
		documentRels("settings.xml")))
	// A documentProtection element in another namespace.
	fx.Raw(t, "testdata", "other-namespace.docx", fx.Docx(fx.Paragraph,
		documentRels("settings.xml"),
		fx.File("word/settings.xml", []byte(`<x:settings xmlns:x="urn:example"><x:documentProtection/></x:settings>`))))
	// A documentProtection element outside the settings part.
	fx.Raw(t, "testdata", "in-document.docx", fx.Docx(protection+fx.Paragraph))
}
