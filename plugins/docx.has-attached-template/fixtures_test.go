package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	relSettings       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"
	strictBase        = "http://purl.oclc.org/ooxml/officeDocument/relationships/"
	strictSettings    = strictBase + "settings"
	strictAttachedTpl = strictBase + "attachedTemplate"
)

// settings is a settings part that references its template relationship.
var settings = fx.WordXML("settings", `<w:attachedTemplate r:id="rId1" xmlns:r="`+
	`http://schemas.openxmlformats.org/officeDocument/2006/relationships"/>`)

// docRels returns the main document's .rels with a settings relationship to
// target.
func docRels(typ, target string) fx.Entry {
	return fx.File("word/_rels/document.xml.rels", fx.Rels([3]string{typ, target, ""}))
}

// tplRels returns the .rels part name with an external template
// relationship of type typ.
func tplRels(name, typ string) fx.Entry {
	return fx.File(name, fx.Rels([3]string{typ, "file:///C:/templates/sample.dotx", "External"}))
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "attached.docx", fx.Docx(fx.Paragraph,
		docRels(relSettings, "settings.xml"),
		fx.File("word/settings.xml", settings),
		tplRels("word/_rels/settings.xml.rels", fx.RelAttachedTmpl)))
	// Strict relationship types and an absolute settings target.
	fx.Raw(t, "testdata", "strict.docx", fx.Docx(fx.Paragraph,
		docRels(strictSettings, "/word/settings.xml"),
		fx.File("word/settings.xml", settings),
		tplRels("word/_rels/settings.xml.rels", strictAttachedTpl)))
	// The settings part under another name and case.
	fx.Raw(t, "testdata", "renamed.docx", fx.Docx(fx.Paragraph,
		docRels(relSettings, "conf/Prefs.xml"),
		fx.File("WORD/CONF/PREFS.XML", settings),
		tplRels("word/conf/_rels/prefs.xml.rels", fx.RelAttachedTmpl)))
	// A settings part with no .rels part.
	fx.Raw(t, "testdata", "settings-no-rels.docx", fx.Docx(fx.Paragraph,
		docRels(relSettings, "settings.xml"),
		fx.File("word/settings.xml", fx.WordXML("settings", ""))))
	// Settings relationships, but none to a template.
	fx.Raw(t, "testdata", "other-rel.docx", fx.Docx(fx.Paragraph,
		docRels(relSettings, "settings.xml"),
		fx.File("word/settings.xml", settings),
		tplRels("word/_rels/settings.xml.rels", fx.RelHyperlink)))
	// A template relationship on the main document, not on the settings.
	fx.Raw(t, "testdata", "on-document.docx", fx.Docx(fx.Paragraph,
		tplRels("word/_rels/document.xml.rels", fx.RelAttachedTmpl)))
	// A template relationship on a settings part nothing points to.
	fx.Raw(t, "testdata", "orphan-settings.docx", fx.Docx(fx.Paragraph,
		fx.File("word/settings.xml", settings),
		tplRels("word/_rels/settings.xml.rels", fx.RelAttachedTmpl)))
	// The settings relationship dangles, its .rels part alone remains.
	fx.Raw(t, "testdata", "dangling-settings.docx", fx.Docx(fx.Paragraph,
		docRels(relSettings, "settings.xml"),
		tplRels("word/_rels/settings.xml.rels", fx.RelAttachedTmpl)))
}
