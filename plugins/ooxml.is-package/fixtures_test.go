package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// relStrictOfficeDocument is the ISO strict officeDocument relationship type.
const relStrictOfficeDocument = "http://purl.oclc.org/ooxml/officeDocument/relationships/officeDocument"

func writeFixtures(t *testing.T) {
	types := fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain}))
	document := fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph))
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	fx.Raw(t, "testdata", "strict.docx", fx.Zip(
		types,
		fx.File("_rels/.rels", fx.Rels([3]string{relStrictOfficeDocument, "/word/document.xml", ""})),
		document,
	))
	fx.Raw(t, "testdata", "plain.zip", fx.Zip(fx.File("a.txt", fx.Sample)))
	fx.Raw(t, "testdata", "no-root-rels.docx", fx.Zip(types, document))
	fx.Raw(t, "testdata", "bad-root-rels.docx", fx.Zip(
		types,
		fx.File("_rels/.rels", []byte(`<Relationships><Relationship`)),
		document,
	))
	fx.Raw(t, "testdata", "missing-root-part.docx", fx.Zip(
		types,
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "word/document.xml", ""})),
	))
	fx.Raw(t, "testdata", "external-target.docx", fx.Zip(
		types,
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "word/document.xml", "External"})),
		document,
	))
	fx.Raw(t, "testdata", "no-office-document.docx", fx.Zip(
		types,
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelHyperlink, "word/document.xml", ""})),
		document,
	))
}
