package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "whitespace.json", []byte(" \n\t\r\n"))
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "object.json", []byte("{\"id\": 1, \"city\": \"Liège\", \"tags\": [\"a\", {\"b\": null}], \"ok\": true}\n"))
	fx.Raw(t, "testdata", "array.json", []byte("[1, -2.5e3, \"x\", false, [], {}]"))
	fx.Raw(t, "testdata", "scalar-string.json", []byte("  \"just a string\"  \n"))
	fx.Raw(t, "testdata", "scalar-number.json", []byte("42"))
	fx.Raw(t, "testdata", "bom.json", []byte("\xef\xbb\xbf{\"a\": 1}"))
	fx.Raw(t, "testdata", "two-values.json", []byte("{\"a\": 1}\n{\"a\": 2}\n"))
	fx.Raw(t, "testdata", "two-scalars.json", []byte("1 2"))
	fx.Raw(t, "testdata", "trailing-garbage.json", []byte("[1, 2] x"))
	fx.Raw(t, "testdata", "truncated-array.json", []byte("[1, 2"))
	fx.Raw(t, "testdata", "truncated-object.json", []byte("{\"a\": "))
	fx.Raw(t, "testdata", "missing-colon.json", []byte("{\"a\" 1}"))
	fx.Raw(t, "testdata", "trailing-comma.json", []byte("[1, 2,]"))
	fx.Raw(t, "testdata", "unquoted-key.json", []byte("{a: 1}"))
	fx.Raw(t, "testdata", "invalid-utf8.json", []byte("{\"a\": \"\xff\"}"))
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
}
