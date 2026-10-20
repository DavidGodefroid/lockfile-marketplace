package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk/harness"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestConformance(t *testing.T) {
	writeFixtures(t)
	harness.Check(t, ".")
}

func TestNoAnswer(t *testing.T) {
	gpsPNG := fx.PNG(4, 4, exif(true))
	gpsJPEG := fx.JPEG(8, 8, fx.ExifAPP1(fx.TIFF(ii, true)))
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":             nil,
		"sample.txt":        fx.Sample,
		"sample.docx":       fx.Docx(fx.Paragraph),
		"png magic only":    gpsPNG[:8],
		"png no IEND":       gpsPNG[:len(gpsPNG)-12],
		"png bad tiff":      fx.PNG(4, 4, fx.Chunk{Type: "eXIf", Data: []byte("XX\x00\x2a\x00\x00\x00\x08")}),
		"png huge eXIf":     fx.PNG(4, 4, fx.Chunk{Type: "eXIf", Data: make([]byte, maxExif+1)}),
		"jpeg cut in exif":  gpsJPEG[:20],
		"jpeg short tiff":   fx.JPEG(8, 8, fx.ExifAPP1([]byte("MM\x00\x2a"))),
		"jpeg bad length":   append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x01}, make([]byte, 8)...),
		"jpeg not a marker": []byte{0xFF, 0xD8, 0xFF, 0xD8, 0x00},
	})
}

func TestTIFFBounds(t *testing.T) {
	ok := fx.TIFF(binary.BigEndian, true)
	badOffset := bytes.Clone(ok)
	binary.BigEndian.PutUint32(badOffset[4:8], uint32(len(ok)-1))
	badCount := bytes.Clone(ok)
	binary.BigEndian.PutUint16(badCount[8:10], 0xFFFF)
	hugeOffset := bytes.Clone(ok)
	binary.BigEndian.PutUint32(hugeOffset[4:8], 0xFFFFFFFF)
	badMagic := bytes.Clone(ok)
	badMagic[3] = 43
	for name, b := range map[string][]byte{
		"header":      ok[:7],
		"ifd0 offset": badOffset,
		"huge offset": hugeOffset,
		"entry count": badCount,
		"magic":       badMagic,
	} {
		if got, err := tiffGPS(b); err == nil {
			t.Errorf("%s: tiffGPS = %v, nil; want an error", name, got)
		}
	}
	if got, err := tiffGPS(ok); err != nil || !got {
		t.Errorf("tiffGPS(valid) = %v, %v; want true, nil", got, err)
	}
}

func TestJPEGSkipsFillAndRestart(t *testing.T) {
	b := []byte{0xFF, 0xD8, 0xFF, 0xFF, 0xD0}
	b = append(b, fx.ExifAPP1(fx.TIFF(binary.LittleEndian, true))...)
	b = append(b, 0xFF, 0xD9)
	if got, err := hasGPS(bytes.NewReader(b)); err != nil || !got {
		t.Errorf("hasGPS = %v, %v; want true, nil", got, err)
	}
}
