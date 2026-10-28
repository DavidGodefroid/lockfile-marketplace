package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// adobe is an APP14 Adobe segment with the given colour transform (0: none,
// CMYK; 2: YCCK), as written by the tools that produce four-component JPEGs.
func adobe(transform byte) []byte {
	return fx.Segment(0xEE, []byte{'A', 'd', 'o', 'b', 'e', 0, 100, 0, 0, 0, 0, transform})
}

// fourComponent returns the headers of a w×h baseline JPEG with four
// components and 1×1 sampling: SOI, the given segments, SOF0, an SOS that
// names all four components, and EOI. jpeg.DecodeConfig stops at the SOS, so
// no entropy-coded data is needed.
func fourComponent(w, h int, segments ...[]byte) []byte {
	b := []byte{0xFF, 0xD8}
	for _, s := range segments {
		b = append(b, s...)
	}
	sof := []byte{8, byte(h >> 8), byte(h), byte(w >> 8), byte(w), 4}
	sos := []byte{4}
	for c := byte(1); c <= 4; c++ {
		sof = append(sof, c, 0x11, 0)
		sos = append(sos, c, 0x00)
	}
	sos = append(sos, 0, 63, 0)
	b = append(b, fx.Segment(0xC0, sof)...)
	b = append(b, fx.Segment(0xDA, sos)...)
	return append(b, 0xFF, 0xD9)
}

// colourJPEG returns a w×h three-component (YCbCr) JPEG.
func colourJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = byte(i)
	}
	img.Set(0, 0, color.RGBA{R: 200, A: 255})
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "cmyk.jpg", fourComponent(64, 48, adobe(0)))
	fx.Raw(t, "testdata", "ycck.jpg", fourComponent(64, 48, adobe(2)))
	fx.Raw(t, "testdata", "four-component-no-adobe.jpg", fourComponent(16, 16))
	fx.Raw(t, "testdata", "gray.jpg", fx.JPEG(64, 48))
	fx.Raw(t, "testdata", "ycbcr.jpg", colourJPEG(64, 48))
}
