package main

import (
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// late puts the only NUL byte past the first 32 KiB read chunk.
var late = append(bytes.Repeat([]byte("a"), 40*1024), 0)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "high-bytes.bin", []byte{0x01, 0x7f, 0x80, 0xff, 0xfe})
	fx.Raw(t, "testdata", "nul-only.bin", []byte{0})
	fx.Raw(t, "testdata", "nul-first.bin", append([]byte{0}, fx.Sample...))
	fx.Raw(t, "testdata", "nul-middle.txt", []byte("sample\x00text\n"))
	fx.Raw(t, "testdata", "nul-last.txt", []byte("sample text\n\x00"))
	fx.Raw(t, "testdata", "late-nul.bin", late)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
}
