package main

import (
	"fmt"
	"strings"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const head = "From: a@example.com\r\nTo: b@example.com\r\nSubject: sample\r\nMIME-Version: 1.0\r\n"

// eml builds a message with the given extra header lines and body.
func eml(header, body string) []byte {
	return []byte(head + header + "\r\n" + body)
}

// multi builds a multipart body with boundary b and the given raw parts
// (each a header block, a blank line and a body).
func multi(b string, parts ...string) string {
	var s strings.Builder
	for _, p := range parts {
		fmt.Fprintf(&s, "--%s\r\n%s\r\n", b, p)
	}
	fmt.Fprintf(&s, "--%s--\r\n", b)
	return s.String()
}

func ct(v string) string { return "Content-Type: " + v + "\r\n" }

func cd(v string) string { return "Content-Disposition: " + v + "\r\n" }

// nested wraps a text part in n multipart/mixed levels.
func nested(n int) []byte {
	body := ct("text/plain") + "\r\nsample\r\n"
	for i := n; i > 0; i-- {
		b := fmt.Sprintf("b%d", i)
		body = ct("multipart/mixed; boundary="+b) + "\r\n" + multi(b, body)
	}
	return []byte(head + body)
}

var (
	textPart   = ct("text/plain; charset=utf-8") + "\r\nsample\r\n"
	htmlPart   = ct("text/html") + "\r\n<p>sample</p>\r\n"
	attachPart = ct("application/octet-stream") + cd("attachment; filename=\"sample.bin\"") + "\r\nAAAA\r\n"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "plain.eml", eml(ct("text/plain"), "sample\r\n"))
	fx.Raw(t, "testdata", "no-content-type.eml", eml("", "sample\r\n"))
	fx.Raw(t, "testdata", "alternative.eml", eml(ct("multipart/alternative; boundary=x"), multi("x", textPart, htmlPart)))
	fx.Raw(t, "testdata", "mixed-attachment.eml", eml(ct("multipart/mixed; boundary=x"), multi("x", textPart, attachPart)))
	fx.Raw(t, "testdata", "nested-attachment.eml", eml(ct("multipart/mixed; boundary=outer"), multi("outer",
		ct("multipart/alternative; boundary=inner")+"\r\n"+multi("inner", textPart, htmlPart),
		attachPart)))
	fx.Raw(t, "testdata", "text-attachment.eml", eml(ct("multipart/mixed; boundary=x"), multi("x", textPart,
		ct("text/plain")+cd("attachment")+"\r\nsample\r\n")))
	fx.Raw(t, "testdata", "inline-image-filename.eml", eml(ct("multipart/related; boundary=x"), multi("x", htmlPart,
		ct("image/png")+cd("inline; filename=\"sample.png\"")+"\r\nAAAA\r\n")))
	fx.Raw(t, "testdata", "rfc2231-filename.eml", eml(ct("multipart/mixed; boundary=x"), multi("x", textPart,
		ct("application/pdf")+cd("inline; filename*=utf-8''sample.pdf")+"\r\nAAAA\r\n")))
	fx.Raw(t, "testdata", "inline-image-no-filename.eml", eml(ct("multipart/related; boundary=x"), multi("x", htmlPart,
		ct("image/png")+cd("inline")+"\r\nAAAA\r\n")))
	fx.Raw(t, "testdata", "inline-text-filename.eml", eml(ct("multipart/mixed; boundary=x"), multi("x",
		ct("text/plain")+cd("inline; filename=\"sample.txt\"")+"\r\nsample\r\n")))
	fx.Raw(t, "testdata", "single-part-attachment.eml", eml(ct("application/pdf")+cd("attachment"), "AAAA\r\n"))
	fx.Raw(t, "testdata", "depth-32.eml", nested(maxDepth))
	fx.Raw(t, "testdata", "wrapper-boundary-in-body.eml", eml(ct("multipart/mixed; boundary=x"), multi("x",
		ct("text/plain")+"\r\n--lockfile-eml\r\n--lockfile-eml-0--\r\n", attachPart)))
}
