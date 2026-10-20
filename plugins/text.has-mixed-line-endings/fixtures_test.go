package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "no-newline.txt", []byte("sample"))
	fx.Raw(t, "testdata", "lf.txt", []byte("sample\nsample\n"))
	fx.Raw(t, "testdata", "crlf.txt", []byte("sample\r\nsample\r\n"))
	fx.Raw(t, "testdata", "cr.txt", []byte("sample\rsample\r"))
	fx.Raw(t, "testdata", "crlf-blank-lines.txt", []byte("sample\r\n\r\n\r\nsample"))
	fx.Raw(t, "testdata", "cr-blank-lines.txt", []byte("sample\r\r\rsample"))
	fx.Raw(t, "testdata", "crlf-lf.txt", []byte("sample\r\nsample\n"))
	fx.Raw(t, "testdata", "crlf-cr.txt", []byte("sample\r\nsample\rsample"))
	fx.Raw(t, "testdata", "lf-cr.txt", []byte("sample\nsample\r"))
	fx.Raw(t, "testdata", "lf-crlf-reversed.txt", []byte("sample\n\rsample"))
	fx.Raw(t, "testdata", "trailing-cr.txt", []byte("sample\r\nsample\r"))
}
