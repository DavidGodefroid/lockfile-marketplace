package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/gif"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// sized returns fx.PNG(16, 16) with the IHDR width and height rewritten to
// w and h and the IHDR CRC fixed, so the header parses but the pixel data
// does not match it.
func sized(w, h uint32) []byte {
	b := fx.PNG(16, 16)
	const ihdr = 8 + 4 // signature, chunk length
	binary.BigEndian.PutUint32(b[ihdr+4:], w)
	binary.BigEndian.PutUint32(b[ihdr+8:], h)
	binary.BigEndian.PutUint32(b[ihdr+4+13:], crc32.ChecksumIEEE(b[ihdr:ihdr+4+13]))
	return b
}

func gifImage() []byte {
	var buf bytes.Buffer
	if err := gif.Encode(&buf, image.NewGray(image.Rect(0, 0, 16, 16)), nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func writeFixtures(t *testing.T) {
	png, jpg := fx.PNG(64, 48), fx.JPEG(64, 48)
	fx.Raw(t, "testdata", "image.png", png)
	fx.Raw(t, "testdata", "image.jpg", jpg)
	fx.Raw(t, "testdata", "truncated.png", png[:len(png)/2])
	fx.Raw(t, "testdata", "truncated.jpg", jpg[:len(jpg)/2])
	fx.Raw(t, "testdata", "header-only.png", png[:33])
	fx.Raw(t, "testdata", "at-pixel-limit.png", sized(8000, 4000))
	fx.Raw(t, "testdata", "image.gif", gifImage())
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "empty.bin", nil)
}
