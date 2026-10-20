package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "utf8.txt", append([]byte{0xEF, 0xBB, 0xBF}, "hello\n"...))
	fx.Raw(t, "testdata", "utf8-only.txt", []byte{0xEF, 0xBB, 0xBF})
	fx.Raw(t, "testdata", "utf8-truncated.txt", []byte{0xEF, 0xBB})
	fx.Raw(t, "testdata", "utf16be.txt", []byte{0xFE, 0xFF, 0x00, 'h', 0x00, 'i'})
	fx.Raw(t, "testdata", "utf16le.txt", []byte{0xFF, 0xFE, 'h', 0x00, 'i', 0x00})
	fx.Raw(t, "testdata", "utf32le.txt", []byte{0xFF, 0xFE, 0x00, 0x00, 'h', 0x00, 0x00, 0x00})
	fx.Raw(t, "testdata", "utf32be.txt", []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 'h'})
	fx.Raw(t, "testdata", "utf32be-truncated.txt", []byte{0x00, 0x00, 0xFE})
	fx.Raw(t, "testdata", "one-byte.txt", []byte{0xFE})
	fx.Raw(t, "testdata", "bom-not-at-start.txt", append([]byte("x"), 0xEF, 0xBB, 0xBF))
}
