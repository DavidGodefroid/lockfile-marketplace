package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// Thumbnail relationship types, as written by OPC producers and in the
// strict relationships base.
const (
	relThumbPackage = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail"
	relThumbStrict  = "http://purl.oclc.org/ooxml/officeDocument/relationships/metadata/thumbnail"
)

// docx returns a minimal DOCX whose root relationships are the main
// document plus rels, with extra entries.
func docx(rels [][3]string, extra ...fx.Entry) []byte {
	all := append([][3]string{{fx.RelOfficeDocument, "word/document.xml", ""}}, rels...)
	entries := []fx.Entry{
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
		fx.File("_rels/.rels", fx.Rels(all...)),
		fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph)),
	}
	return fx.Zip(append(entries, extra...)...)
}

func writeFixtures(t *testing.T) {
	thumb := fx.File("docProps/thumbnail.jpeg", fx.JPEG(4, 4))
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	fx.Raw(t, "testdata", "thumbnail.docx", docx([][3]string{{relThumbPackage, "docProps/thumbnail.jpeg", ""}}, thumb))
	// A target with a leading slash and different case still names the
	// part, as OPC part names are case-insensitive.
	fx.Raw(t, "testdata", "thumbnail-absolute.docx", docx([][3]string{{relThumbPackage, "/DocProps/Thumbnail.JPEG", ""}}, thumb))
	fx.Raw(t, "testdata", "thumbnail-strict.docx", docx([][3]string{{relThumbStrict, "docProps/thumbnail.jpeg", ""}}, thumb))
	// A thumbnail relationship to a part missing from the package.
	fx.Raw(t, "testdata", "dangling.docx", docx([][3]string{{relThumbPackage, "docProps/thumbnail.jpeg", ""}}))
	// A thumbnail relationship to an external resource.
	fx.Raw(t, "testdata", "external.docx", docx([][3]string{{relThumbPackage, "https://example.invalid/thumb.jpeg", "External"}}))
	// A thumbnail part with no root relationship to it.
	fx.Raw(t, "testdata", "unrelated-part.docx", docx(nil, thumb))
	// A package with no root relationships part at all.
	fx.Raw(t, "testdata", "no-root-rels.docx", fx.Zip(
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
		fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph))))
}
