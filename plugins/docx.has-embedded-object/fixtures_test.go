package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// Relationship types of embedded objects, transitional and strict.
const (
	relOLE       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/oleObject"
	relPackage   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/package"
	relOLEStrict = "http://purl.oclc.org/ooxml/officeDocument/relationships/oleObject"
	relImage     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
)

// docRels is the relationships part of word/document.xml.
func docRels(rels ...[3]string) fx.Entry {
	return fx.File("word/_rels/document.xml.rels", fx.Rels(rels...))
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "ole.docx", fx.Docx(fx.Paragraph,
		docRels([3]string{relOLE, "embeddings/oleObject1.bin", ""}),
		fx.File("word/embeddings/oleObject1.bin", fx.Sample)))
	fx.Raw(t, "testdata", "package.docx", fx.Docx(fx.Paragraph,
		docRels([3]string{relImage, "media/image1.png", ""}, [3]string{relPackage, "embeddings/sample.xlsx", ""}),
		fx.File("word/embeddings/sample.xlsx", fx.Xlsx(nil))))
	fx.Raw(t, "testdata", "strict.docx", fx.Docx(fx.Paragraph,
		docRels([3]string{relOLEStrict, "embeddings/oleObject1.bin", ""})))
	// The relationships part with an upper-case name.
	fx.Raw(t, "testdata", "upper-case.docx", fx.Docx(fx.Paragraph,
		fx.File("WORD/_RELS/DOCUMENT.XML.RELS", fx.Rels([3]string{relOLE, "embeddings/oleObject1.bin", ""}))))
	// Other relationships only.
	fx.Raw(t, "testdata", "image-only.docx", fx.Docx(fx.Paragraph,
		docRels([3]string{relImage, "media/image1.png", ""}, [3]string{fx.RelHyperlink, "https://example.invalid/", "External"})))
	// A linked object points outside the package.
	fx.Raw(t, "testdata", "linked.docx", fx.Docx(fx.Paragraph,
		docRels([3]string{relOLE, "file:///sample.bin", "External"})))
	// An object relationship of another part than the main document.
	fx.Raw(t, "testdata", "header-object.docx", fx.Docx(fx.Paragraph,
		fx.File("word/_rels/header1.xml.rels", fx.Rels([3]string{relOLE, "embeddings/oleObject1.bin", ""}))))
}
