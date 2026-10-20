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
		"empty.bin":          nil,
		"sample.txt":         fx.Sample,
		"plain.zip":          fx.Zip(fx.File("a.txt", fx.Sample)),
		"xlsx":               fx.Xlsx(nil),
		"malformed document": fx.Docx(`<w:body><w:p>`),
		"malformed comments": fx.Docx(fx.Paragraph,
			fx.File("word/comments.xml", []byte(`<w:comments xmlns:w="`+fx.WNS+`"><w:comment`))),
		"comment then malformed part": fx.Docx(fx.Paragraph,
			fx.File("word/comments.xml", fx.WordXML("comments", comment)),
			fx.File("word/settings.xml", []byte(`<w:settings xmlns:w="`+fx.WNS+`"><w:zoom`))),
	})
}
