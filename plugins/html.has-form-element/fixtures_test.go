package main

import (
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// split puts "<fo" at the end of the first 4096-byte bufio buffer and "rm>"
// at the start of the second.
var split = append(bytes.Repeat([]byte("a"), 4096-3), "<form>"...)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "plain.html", []byte("<!DOCTYPE html><html><body><p>hi</p></body></html>\n"))
	fx.Raw(t, "testdata", "form.html", []byte("<html><body><form><p>sample</p></form></body></html>\n"))
	fx.Raw(t, "testdata", "form-upper-attr.html", []byte("<HTML><FORM\tmethod=\"post\"></FORM></HTML>\n"))
	fx.Raw(t, "testdata", "form-self-closing.html", []byte("<div><form/></div>\n"))
	fx.Raw(t, "testdata", "form-newline.html", []byte("<FoRm\naction=\"a\"></form>\n"))
	fx.Raw(t, "testdata", "form-in-comment.html", []byte("<p>x</p><!-- <form> -->\n"))
	fx.Raw(t, "testdata", "form-after-lt.html", []byte("a <<form>\n"))
	fx.Raw(t, "testdata", "split-buffer.html", split)
	fx.Raw(t, "testdata", "format.html", []byte("<format>not a form</format>\n"))
	fx.Raw(t, "testdata", "form-at-eof.html", []byte("<p>x</p><form"))
	fx.Raw(t, "testdata", "form-attr.html", []byte("<button form=\"f\" formaction=\"a\">go</button>\n"))
	fx.Raw(t, "testdata", "form-dash.html", []byte("<form-x></form-x>\n"))
}
