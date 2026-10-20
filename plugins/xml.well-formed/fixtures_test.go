package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "minimal.xml", []byte("<a/>"))
	fx.Raw(t, "testdata", "document.xml", []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"+
		"<!DOCTYPE r>\n<!-- head -->\n<?pi data?>\n<r a=\"1\"><b>t&amp;x&#233;</b><![CDATA[<raw>]]><c/></r>\n<!-- tail -->\n"))
	fx.Raw(t, "testdata", "word-document.xml", fx.WordXML("document", fx.Paragraph))
	fx.Raw(t, "testdata", "us-ascii.xml", []byte("<?xml version=\"1.0\" encoding=\"US-ASCII\"?><r>plain</r>"))
	fx.Raw(t, "testdata", "latin1.xml", []byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><r>Li\xe8ge</r>"))
	fx.Raw(t, "testdata", "latin1-alias.xml", []byte("<?xml version=\"1.0\" encoding=\"latin1\"?><r>\xe9</r>"))
	fx.Raw(t, "testdata", "unsupported-charset.xml", []byte("<?xml version=\"1.0\" encoding=\"windows-1252\"?><r/>"))
	fx.Raw(t, "testdata", "two-roots.xml", []byte("<a/><b/>"))
	fx.Raw(t, "testdata", "text-before-root.xml", []byte("hello<a/>"))
	fx.Raw(t, "testdata", "text-after-root.xml", []byte("<a/>tail"))
	fx.Raw(t, "testdata", "directive-after-root.xml", []byte("<a/><!DOCTYPE a>"))
	fx.Raw(t, "testdata", "late-declaration.xml", []byte("<!-- c --><?xml version=\"1.0\"?><a/>"))
	fx.Raw(t, "testdata", "leading-space-declaration.xml", []byte(" <?xml version=\"1.0\"?><a/>"))
	fx.Raw(t, "testdata", "unclosed.xml", []byte("<a><b></b>"))
	fx.Raw(t, "testdata", "mismatched.xml", []byte("<a><b></a></b>"))
	fx.Raw(t, "testdata", "undefined-entity.xml", []byte("<a>&nbsp;</a>"))
	fx.Raw(t, "testdata", "invalid-utf8.xml", []byte("<a>\xff</a>"))
	fx.Raw(t, "testdata", "bom.xml", []byte("\xef\xbb\xbf<?xml version=\"1.0\" encoding=\"UTF-8\"?><a/>"))
	fx.Raw(t, "testdata", "bom-no-declaration.xml", []byte("\xef\xbb\xbf<a/>"))
	fx.Raw(t, "testdata", "double-bom.xml", []byte("\xef\xbb\xbf\xef\xbb\xbf<a/>"))
	fx.Raw(t, "testdata", "bom-after-root.xml", []byte("<a/>\xef\xbb\xbf"))
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
}
