package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

var (
	// profile stands in for ICC profile data: never parsed, only its
	// container is looked at.
	profile = []byte("sample")
	iccp    = fx.Chunk{Type: "iCCP", Data: append([]byte("sample\x00\x00"), profile...)}
	text    = fx.Chunk{Type: "tEXt", Data: []byte("Comment\x00ICC_PROFILE")}
	iccAPP2 = fx.Segment(0xE2, append([]byte("ICC_PROFILE\x00\x01\x01"), profile...))
	fpxAPP2 = fx.Segment(0xE2, []byte("FPXR\x00\x00sample"))
	iccAPP1 = fx.Segment(0xE1, append([]byte("ICC_PROFILE\x00\x01\x01"), profile...))
	com     = fx.Segment(0xFE, []byte("ICC_PROFILE\x00"))
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "plain.png", fx.PNG(4, 4))
	fx.Raw(t, "testdata", "iccp.png", fx.PNG(4, 4, iccp))
	fx.Raw(t, "testdata", "text-then-iccp.png", fx.PNG(4, 4, text, iccp))
	fx.Raw(t, "testdata", "text-only.png", fx.PNG(4, 4, text))
	fx.Raw(t, "testdata", "plain.jpg", fx.JPEG(8, 8))
	fx.Raw(t, "testdata", "icc.jpg", fx.JPEG(8, 8, iccAPP2))
	fx.Raw(t, "testdata", "fpxr-then-icc.jpg", fx.JPEG(8, 8, fpxAPP2, iccAPP2))
	fx.Raw(t, "testdata", "fpxr-only.jpg", fx.JPEG(8, 8, fpxAPP2))
	fx.Raw(t, "testdata", "icc-id-in-app1.jpg", fx.JPEG(8, 8, iccAPP1, com))
}
