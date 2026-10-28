package main

import (
	"encoding/binary"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

var (
	plainPNG  = fx.PNG(4, 4)
	plainJPEG = fx.JPEG(8, 8)
	// An Exif APP1 whose payload ends with a whole JPEG (a thumbnail), so the
	// segment holds an EOI marker that is not the end of the file.
	thumbAPP1 = fx.ExifAPP1(append(fx.TIFF(binary.BigEndian, false), plainJPEG...))
	// A tEXt chunk whose data spells a full IEND chunk.
	iendText = fx.Chunk{Type: "tEXt", Data: append([]byte("Comment\x00"), plainPNG[len(plainPNG)-12:]...)}
)

func cat(parts ...[]byte) []byte {
	var b []byte
	for _, p := range parts {
		b = append(b, p...)
	}
	return b
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "plain.png", plainPNG)
	fx.Raw(t, "testdata", "iend-in-text.png", fx.PNG(4, 4, iendText))
	fx.Raw(t, "testdata", "appended.png", cat(plainPNG, fx.Sample))
	fx.Raw(t, "testdata", "one-byte.png", cat(plainPNG, []byte{0}))
	fx.Raw(t, "testdata", "zip-polyglot.png", cat(plainPNG, fx.Zip(fx.File("sample.txt", fx.Sample))))
	fx.Raw(t, "testdata", "plain.jpg", plainJPEG)
	fx.Raw(t, "testdata", "thumbnail.jpg", fx.JPEG(8, 8, thumbAPP1))
	fx.Raw(t, "testdata", "appended.jpg", cat(plainJPEG, fx.Sample))
	fx.Raw(t, "testdata", "thumbnail-appended.jpg", cat(fx.JPEG(8, 8, thumbAPP1), []byte{0}))
	fx.Raw(t, "testdata", "zip-polyglot.jpg", cat(plainJPEG, fx.Zip(fx.File("sample.txt", fx.Sample))))
}
