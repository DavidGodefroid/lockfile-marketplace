package main

import (
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// boundary is the size of the read buffer in valid.
const boundary = 64 << 10

// straddle returns ASCII padding followed by r, placed so that the 64 KiB
// buffer boundary falls inside r's encoding.
func straddle(r string) []byte {
	b := bytes.Repeat([]byte("a"), boundary-1)
	return append(b, r...)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "multilingual.txt", []byte("Liège, Gent, Köln, Ελλάδα, 日本, 🙂\n"))
	fx.Raw(t, "testdata", "bom.txt", []byte("\xef\xbb\xbfhello\n"))
	fx.Raw(t, "testdata", "split-3byte.txt", straddle("€ tail\n"))
	fx.Raw(t, "testdata", "split-4byte.txt", straddle("🙂 tail\n"))
	fx.Raw(t, "testdata", "latin1.txt", []byte("Li\xe8ge\n"))
	fx.Raw(t, "testdata", "stray-continuation.txt", []byte("a\x80b"))
	fx.Raw(t, "testdata", "overlong.txt", []byte("\xc0\xaf"))
	fx.Raw(t, "testdata", "surrogate.txt", []byte("\xed\xa0\x80"))
	fx.Raw(t, "testdata", "byte-ff.txt", []byte("ok\xff"))
	fx.Raw(t, "testdata", "truncated-rune.txt", []byte("tail \xe2\x82"))
	fx.Raw(t, "testdata", "invalid-after-boundary.txt", straddle("\xe2\x82 tail\n"))
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
}
