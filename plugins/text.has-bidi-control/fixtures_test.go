package main

import (
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// boundary is the size of the read buffer in hasBidiControl.
const boundary = 64 << 10

// straddle returns ASCII padding followed by s, placed so that the 64 KiB
// buffer boundary falls inside the encoding of s's first rune.
func straddle(s string) []byte {
	b := bytes.Repeat([]byte("a"), boundary-1)
	return append(b, s...)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "multilingual.txt", []byte("sample, Ελλάδα, עברית, العربية, 日本, 🙂\n"))
	fx.Raw(t, "testdata", "mark-only.txt", []byte("sample \u200f\u200e\u061c text\n"))
	fx.Raw(t, "testdata", "neighbours.txt", []byte("sample \u2029\u202f\u2065\u206a text\n"))
	fx.Raw(t, "testdata", "lre.txt", []byte("sample \u202a text \u202c\n"))
	fx.Raw(t, "testdata", "rlo.txt", []byte("sample \u202e text\n"))
	fx.Raw(t, "testdata", "lri.txt", []byte("sample \u2066 text \u2069\n"))
	fx.Raw(t, "testdata", "pdi-only.txt", []byte("sample text \u2069"))
	fx.Raw(t, "testdata", "split-rlo.txt", straddle("\u202e tail\n"))
}

// noAnswer holds inputs that are not valid UTF-8, including one with a bidi
// control before the invalid byte.
var noAnswer = map[string][]byte{
	"latin1.txt":         []byte("sample Li\xe8ge\n"),
	"rlo-then-ff.txt":    []byte("sample \u202e text\xff"),
	"truncated-rune.txt": []byte("sample \xe2\x80"),
	"invalid-after-boundary.txt": append(straddle("\u202e"), []byte(
		" \xc0\xaf tail\n")...),
}
