package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "minimal.xml", []byte("<a/>"))
	fx.Raw(t, "testdata", "word-document.xml", fx.WordXML("document", fx.Paragraph))
	fx.Raw(t, "testdata", "doctype-only.xml", []byte("<!DOCTYPE r>\n<r/>"))
	fx.Raw(t, "testdata", "external-dtd.xml", []byte("<!DOCTYPE r SYSTEM \"sample.dtd\">\n<r/>"))
	fx.Raw(t, "testdata", "internal-entity.xml", []byte("<?xml version=\"1.0\"?>\n"+
		"<!DOCTYPE r [\n<!ENTITY s \"sample\">\n<!ELEMENT r (#PCDATA)>\n"+
		"<!ATTLIST r a CDATA \"x>y\">\n]>\n<r>&s;</r>"))
	fx.Raw(t, "testdata", "system-entity.xml", []byte("<?xml version=\"1.0\"?>\n"+
		"<!DOCTYPE r [\n<!ENTITY s SYSTEM \"sample.txt\">\n]>\n<r>&s;</r>"))
	fx.Raw(t, "testdata", "public-entity.xml", []byte("<!DOCTYPE r [<!ENTITY s PUBLIC \"-//Sample//EN\" 'sample.txt'>]><r/>"))
	fx.Raw(t, "testdata", "parameter-entity.xml", []byte("<!DOCTYPE r [<!ENTITY % p SYSTEM \"sample.dtd\"> %p;]><r/>"))
	fx.Raw(t, "testdata", "unparsed-entity.xml", []byte("<!DOCTYPE r [<!NOTATION n SYSTEM \"sample\">"+
		"<!ENTITY u SYSTEM \"sample.bin\" NDATA n>]><r/>"))
	fx.Raw(t, "testdata", "entity-in-comment.xml", []byte("<!DOCTYPE r [<!-- <!ENTITY s SYSTEM \"sample.txt\"> -->]><r/>"))
	fx.Raw(t, "testdata", "entity-in-literal.xml", []byte("<!DOCTYPE r [<!ENTITY s \"<!ENTITY t SYSTEM 'sample.txt'>\">]><r/>"))
	fx.Raw(t, "testdata", "entity-in-pi.xml", []byte("<!DOCTYPE r [<?pi <!ENTITY s SYSTEM 'sample.txt'> ?>]><r/>"))
	fx.Raw(t, "testdata", "latin1-entity.xml", []byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>"+
		"<!DOCTYPE r [<!ENTITY s SYSTEM \"Li\xe8ge.txt\">]><r>&s;</r>"))
	fx.Raw(t, "testdata", "bom-entity.xml", []byte("\xef\xbb\xbf<!DOCTYPE r [<!ENTITY s SYSTEM \"sample.txt\">]><r/>"))
}
