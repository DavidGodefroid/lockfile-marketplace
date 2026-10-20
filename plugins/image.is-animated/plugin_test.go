package main

import (
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
	apng := fx.PNG(4, 4, actl)
	anim := gifFrames(2)
	jpg := fx.JPEG(8, 8)
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":          nil,
		"sample.txt":     fx.Sample,
		"sample.docx":    fx.Docx(fx.Paragraph),
		"png magic only": apng[:8],
		"png no IEND":    apng[:len(apng)-12],
		"png cut chunk":  apng[:40],
		"gif magic only": anim[:6],
		"gif truncated":  anim[:len(anim)-10],
		"jpeg magic":     jpg[:3],
		"jpeg truncated": jpg[:20],
	})
}
