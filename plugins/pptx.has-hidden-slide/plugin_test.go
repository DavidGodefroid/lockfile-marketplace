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
		"empty.bin":  nil,
		"sample.txt": fx.Sample,
		"plain.zip":  fx.Zip(fx.File("a.txt", fx.Sample)),
		"docx":       fx.Docx(fx.Paragraph),
		"xlsx":       fx.Xlsx(nil),
		// The root relationship targets a part that is absent.
		"missing presentation": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/ppt/presentation.xml", presMain})),
			fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "ppt/presentation.xml", ""}))),
		"malformed root rels": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("_rels/.rels", []byte(`<Relationships><Relationship`))),
		"malformed presentation rels": pptx(presMain, nil,
			fx.File("ppt/_rels/presentation.xml.rels", []byte(`<Relationships><Relationship`))),
		"missing slide":      pptx(presMain, [][3]string{{relSlide, "slides/slide1.xml", ""}}),
		"malformed slide":    oneSlide([]byte(`<p:sld xmlns:p="` + pNS + `" show="0"><p:cSld>`)),
		"slide root not sld": oneSlide([]byte(`<p:notes xmlns:p="` + pNS + `"/>`)),
		"empty slide part":   oneSlide(nil),
		"hidden then malformed slide": pptx(presMain,
			[][3]string{{relSlide, "slides/slide1.xml", ""}, {relSlide, "slides/slide2.xml", ""}},
			fx.File("ppt/slides/slide1.xml", slide(pNS, ` show="0"`)),
			fx.File("ppt/slides/slide2.xml", []byte(`<p:sld xmlns:p="`+pNS+`"><p:cSld`))),
	})
}
