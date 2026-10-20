package main

import (
	"bytes"
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
	iccPNG := fx.PNG(4, 4, iccp)
	iccJPEG := fx.JPEG(8, 8, iccAPP2)
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":             nil,
		"sample.txt":        fx.Sample,
		"sample.docx":       fx.Docx(fx.Paragraph),
		"png magic only":    iccPNG[:8],
		"png cut in iCCP":   iccPNG[:8+25+12],
		"png no IEND":       iccPNG[:len(iccPNG)-12],
		"jpeg cut in icc":   iccJPEG[:10],
		"jpeg cut in app2":  iccJPEG[:len(iccAPP2)],
		"jpeg bad length":   append([]byte{0xFF, 0xD8, 0xFF, 0xE2, 0x00, 0x01}, make([]byte, 8)...),
		"jpeg not a marker": []byte{0xFF, 0xD8, 0xFF, 0xD8, 0x00},
	})
}

func TestJPEGSkipsFillAndRestart(t *testing.T) {
	b := []byte{0xFF, 0xD8, 0xFF, 0xFF, 0xD0}
	b = append(b, iccAPP2...)
	b = append(b, 0xFF, 0xD9)
	if got, err := hasICC(bytes.NewReader(b)); err != nil || !got {
		t.Errorf("hasICC = %v, %v; want true, nil", got, err)
	}
}

func TestShortAPP2IsNotICC(t *testing.T) {
	b := fx.JPEG(8, 8, fx.Segment(0xE2, []byte("ICC")))
	if got, err := hasICC(bytes.NewReader(b)); err != nil || got {
		t.Errorf("hasICC = %v, %v; want false, nil", got, err)
	}
}
