package main

import (
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// split puts "<ifr" at the end of the first 4096-byte bufio buffer and
// "ame>" at the start of the second.
var split = append(bytes.Repeat([]byte("a"), 4096-4), "<iframe>"...)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "plain.html", []byte("<!DOCTYPE html><html><body><p>hi</p></body></html>\n"))
	fx.Raw(t, "testdata", "iframe.html", []byte("<html><body><iframe></iframe></body></html>\n"))
	fx.Raw(t, "testdata", "iframe-upper-src.html", []byte("<HTML><IFRAME\tsrc=\"a.html\"></IFRAME></HTML>\n"))
	fx.Raw(t, "testdata", "iframe-self-closing.html", []byte("<div><iframe/></div>\n"))
	fx.Raw(t, "testdata", "iframe-newline.html", []byte("<IfRaMe\nwidth=\"1\"></iframe>\n"))
	fx.Raw(t, "testdata", "iframe-in-comment.html", []byte("<p>x</p><!-- <iframe> -->\n"))
	fx.Raw(t, "testdata", "iframe-after-lt.html", []byte("a <<iframe>\n"))
	fx.Raw(t, "testdata", "split-buffer.html", split)
	fx.Raw(t, "testdata", "iframes.html", []byte("<iframes>not an iframe</iframes>\n"))
	fx.Raw(t, "testdata", "iframe-at-eof.html", []byte("<p>x</p><iframe"))
	fx.Raw(t, "testdata", "frame.html", []byte("<frameset><frame src=\"a.html\"></frameset>\n"))
	fx.Raw(t, "testdata", "iframe-dash.html", []byte("<iframe-x></iframe-x><p>iframe</p>\n"))
}
