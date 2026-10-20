package main

import (
	"archive/zip"
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
	enc := fx.File("META-INF/encryption.xml", []byte(encryption))
	stored := func(body string) fx.Entry {
		return fx.Entry{Name: "mimetype", Body: []byte(body), Method: zip.Store}
	}
	badMime := fx.Entry{Name: "mimetype", Body: []byte(mimetype), Method: zip.Store, Raw: true, CRC32: 1}
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt":         fx.Sample,
		"empty":              nil,
		"truncated":          epub(enc)[:20],
		"empty-zip":          fx.Zip(),
		"plain-zip":          fx.Zip(fx.File("a.txt", []byte("sample\n")), enc),
		"docx":               fx.Docx(fx.Paragraph),
		"mimetype-not-first": fx.Zip(fx.File("META-INF/container.xml", []byte(container)), stored(mimetype), enc),
		"other-mimetype":     fx.Zip(stored("application/oebps-package+xml"), enc),
		"mimetype-newline":   fx.Zip(stored(mimetype+"\n"), enc),
		"mimetype-truncated": fx.Zip(stored("application/epub"), enc),
		"mimetype-empty":     fx.Zip(stored(""), enc),
		"mimetype-bad-crc":   fx.Zip(badMime, enc),
	})
}
