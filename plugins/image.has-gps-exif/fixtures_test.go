package main

import (
	"encoding/binary"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

var (
	mm    = binary.BigEndian
	ii    = binary.LittleEndian
	xmp   = fx.Segment(0xE1, []byte("http://ns.adobe.com/xap/1.0/\x00<x:xmpmeta/>"))
	com   = fx.Segment(0xFE, []byte("comment"))
	exif  = func(gps bool) fx.Chunk { return fx.Chunk{Type: "eXIf", Data: fx.TIFF(mm, gps)} }
	other = fx.Chunk{Type: "tEXt", Data: []byte("Comment\x00GPS")}
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "plain.png", fx.PNG(4, 4))
	fx.Raw(t, "testdata", "gps-mm.png", fx.PNG(4, 4, exif(true)))
	fx.Raw(t, "testdata", "gps-ii.png", fx.PNG(4, 4, other, fx.Chunk{Type: "eXIf", Data: fx.TIFF(ii, true)}))
	fx.Raw(t, "testdata", "no-gps.png", fx.PNG(4, 4, exif(false)))
	fx.Raw(t, "testdata", "plain.jpg", fx.JPEG(8, 8))
	fx.Raw(t, "testdata", "gps-ii.jpg", fx.JPEG(8, 8, fx.ExifAPP1(fx.TIFF(ii, true))))
	fx.Raw(t, "testdata", "gps-mm.jpg", fx.JPEG(8, 8, com, fx.ExifAPP1(fx.TIFF(mm, true))))
	fx.Raw(t, "testdata", "no-gps.jpg", fx.JPEG(8, 8, fx.ExifAPP1(fx.TIFF(ii, false))))
	fx.Raw(t, "testdata", "xmp-only.jpg", fx.JPEG(8, 8, xmp))
	fx.Raw(t, "testdata", "xmp-then-gps.jpg", fx.JPEG(8, 8, xmp, fx.ExifAPP1(fx.TIFF(mm, true))))
}
