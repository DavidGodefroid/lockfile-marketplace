package main

import (
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// split puts "<scr" at the end of the first 4096-byte bufio buffer and
// "ipt>" at the start of the second.
var split = append(bytes.Repeat([]byte("a"), 4096-4), "<script>"...)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "plain.html", []byte("<!DOCTYPE html><html><body><p>hi</p></body></html>\n"))
	fx.Raw(t, "testdata", "script.html", []byte("<html><head><script>var a = 1;</script></head></html>\n"))
	fx.Raw(t, "testdata", "script-upper-src.html", []byte("<HTML><SCRIPT\tsrc=\"a.js\"></SCRIPT></HTML>\n"))
	fx.Raw(t, "testdata", "script-self-closing.html", []byte("<svg><script/></svg>\n"))
	fx.Raw(t, "testdata", "script-newline.html", []byte("<ScRiPt\ntype=\"module\"></script>\n"))
	fx.Raw(t, "testdata", "script-in-comment.html", []byte("<p>x</p><!-- <script> -->\n"))
	fx.Raw(t, "testdata", "script-after-lt.html", []byte("a <<script>\n"))
	fx.Raw(t, "testdata", "split-buffer.html", split)
	fx.Raw(t, "testdata", "scripts.html", []byte("<scripts>not a script</scripts>\n"))
	fx.Raw(t, "testdata", "script-at-eof.html", []byte("<p>x</p><script"))
	fx.Raw(t, "testdata", "onclick.html", []byte("<button onclick=\"alert(1)\">go</button>\n"))
	fx.Raw(t, "testdata", "noscript-dash.html", []byte("<script-x></script-x><noscript>x</noscript>\n"))
}
