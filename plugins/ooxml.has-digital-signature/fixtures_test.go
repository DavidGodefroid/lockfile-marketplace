package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// sigXML is a synthetic signature part; its content is never read.
var sigXML = []byte(`<Signature xmlns="http://www.w3.org/2000/09/xmldsig#"/>`)

// rootRels returns package root rels with the main document and, when
// origin is set, a digital-signature origin relationship to it.
func rootRels(origin string) fx.Entry {
	rels := [][3]string{{fx.RelOfficeDocument, "word/document.xml", ""}}
	if origin != "" {
		rels = append(rels, [3]string{relOrigin, origin, ""})
	}
	return fx.File("_rels/.rels", fx.Rels(rels...))
}

// signed returns a docx whose origin part links sigTarget, plus extra.
func signed(sigTarget string, extra ...fx.Entry) []byte {
	entries := []fx.Entry{
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
		rootRels("_xmlsignatures/origin.sigs"),
		fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph)),
		fx.File("_xmlsignatures/origin.sigs", nil),
		fx.File("_xmlsignatures/_rels/origin.sigs.rels", fx.Rels([3]string{relSignature, sigTarget, ""})),
	}
	return fx.Zip(append(entries, extra...)...)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "sample.xlsx", fx.Xlsx(nil))
	// Origin plus one signature part, relative target.
	fx.Raw(t, "testdata", "signed.docx", signed("sig1.xml", fx.File("_xmlsignatures/sig1.xml", sigXML)))
	// Origin plus one signature part, absolute target.
	fx.Raw(t, "testdata", "signed-absolute.docx", signed("/_xmlsignatures/sig1.xml", fx.File("_xmlsignatures/sig1.xml", sigXML)))
	// Origin present but its signature target is missing.
	fx.Raw(t, "testdata", "dangling-signature.docx", signed("sig1.xml"))
	// Origin part with no rels: no signature linked.
	fx.Raw(t, "testdata", "origin-only.docx", fx.Zip(
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/word/document.xml", fx.DocxMain})),
		rootRels("_xmlsignatures/origin.sigs"),
		fx.File("word/document.xml", fx.WordXML("document", fx.Paragraph)),
		fx.File("_xmlsignatures/origin.sigs", nil)))
	// A signature part with no origin relationship in the root rels.
	fx.Raw(t, "testdata", "no-origin.docx", fx.Docx(fx.Paragraph,
		fx.File("_xmlsignatures/origin.sigs", nil),
		fx.File("_xmlsignatures/_rels/origin.sigs.rels", fx.Rels([3]string{relSignature, "sig1.xml", ""})),
		fx.File("_xmlsignatures/sig1.xml", sigXML)))
}
