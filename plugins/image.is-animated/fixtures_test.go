package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/gif"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

var (
	actl  = fx.Chunk{Type: "acTL", Data: binary.BigEndian.AppendUint32(binary.BigEndian.AppendUint32(nil, 2), 0)}
	other = fx.Chunk{Type: "tEXt", Data: []byte("Comment\x00sample")}
)

// gifFrames returns a 4×4 GIF with n frames.
func gifFrames(n int) []byte {
	pal := color.Palette{color.Black, color.White}
	g := &gif.GIF{}
	for i := range n {
		img := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
		img.SetColorIndex(i%4, 0, 1)
		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, 10)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "still.png", fx.PNG(4, 4))
	fx.Raw(t, "testdata", "text-chunk.png", fx.PNG(4, 4, other))
	fx.Raw(t, "testdata", "animated.png", fx.PNG(4, 4, actl))
	fx.Raw(t, "testdata", "still.gif", gifFrames(1))
	fx.Raw(t, "testdata", "two-frames.gif", gifFrames(2))
	fx.Raw(t, "testdata", "three-frames.gif", gifFrames(3))
	fx.Raw(t, "testdata", "still.jpg", fx.JPEG(8, 8))
}
