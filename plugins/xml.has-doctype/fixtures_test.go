package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "minimal.xml", []byte("<a/>"))
	fx.Raw(t, "testdata", "declaration-only.xml", []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!-- sample -->\n<r><b>sample</b></r>\n"))
	fx.Raw(t, "testdata", "word-document.xml", fx.WordXML("document", fx.Paragraph))
	fx.Raw(t, "testdata", "commented-doctype.xml", []byte("<!-- <!DOCTYPE r> --><r/>"))
	fx.Raw(t, "testdata", "cdata-doctype.xml", []byte("<r><![CDATA[<!DOCTYPE r>]]></r>"))
	fx.Raw(t, "testdata", "doctype.xml", []byte("<!DOCTYPE r><r/>"))
	fx.Raw(t, "testdata", "doctype-after-declaration.xml", []byte("<?xml version=\"1.0\"?>\n<!DOCTYPE r>\n<r>sample</r>\n"))
	fx.Raw(t, "testdata", "doctype-system.xml", []byte("<!DOCTYPE r SYSTEM \"sample.dtd\"><r/>"))
	fx.Raw(t, "testdata", "doctype-internal-subset.xml", []byte("<!DOCTYPE r [\n<!ELEMENT r (#PCDATA)>\n<!ATTLIST r a CDATA \"x>y\">\n]>\n<r>sample</r>"))
	fx.Raw(t, "testdata", "bom-doctype.xml", []byte("\xef\xbb\xbf<?xml version=\"1.0\" encoding=\"UTF-8\"?><!DOCTYPE r><r/>"))
	fx.Raw(t, "testdata", "latin1-doctype.xml", []byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><!DOCTYPE r><r>\xe9</r>"))
}
