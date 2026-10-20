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
		"bad content types": fx.Zip(
			fx.File("[Content_Types].xml", []byte(`<Types><Default`)),
			fx.File("_rels/.rels", fx.Rels([3]string{relThumbPackage, "docProps/thumbnail.jpeg", ""})),
			fx.File("docProps/thumbnail.jpeg", fx.Sample),
		),
		"bad root rels": fx.Zip(
			fx.File("[Content_Types].xml", fx.ContentTypes()),
			fx.File("_rels/.rels", []byte(`<Relationships><Relationship`)),
			fx.File("docProps/thumbnail.jpeg", fx.Sample),
		),
	})
}
