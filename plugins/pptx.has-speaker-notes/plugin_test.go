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
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty.bin":   nil,
		"sample.txt":  fx.Sample,
		"plain.zip":   fx.Zip(fx.File("a.txt", fx.Sample)),
		"sample.docx": fx.Docx(fx.Paragraph),
		"sample.xlsx": fx.Xlsx(nil),
		// A package with no root relationships part.
		"no-root-rels.pptx": fx.Zip(fx.File("[Content_Types].xml", fx.ContentTypes())),
		"bad root rels.pptx": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("_rels/.rels", []byte(`<Relationships><Relationship`))),
		"bad slide rels.pptx": pptx(ctPresentation, [][][3]string{nil},
			fx.File("ppt/slides/_rels/slide1.xml.rels", []byte(`<Relationships><Relationship`))),
	})
}
