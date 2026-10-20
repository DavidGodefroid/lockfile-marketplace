package main

import (
	"archive/zip"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// mimetype is the stored first entry of an ODF package of media type mt.
func mimetype(mt string) fx.Entry {
	return fx.Entry{Name: "mimetype", Body: []byte(mt), Method: zip.Store}
}

// odt returns an ODF text package with the extra entries after its
// content.
func odt(extra ...fx.Entry) []byte {
	entries := []fx.Entry{
		mimetype("application/vnd.oasis.opendocument.text"),
		fx.File("content.xml", []byte(`<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"/>`)),
		fx.File("META-INF/manifest.xml", []byte(`<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0"/>`)),
	}
	return fx.Zip(append(entries, extra...)...)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.odt", odt())
	// A Basic library module.
	fx.Raw(t, "testdata", "basic.odt", odt(
		fx.File("Basic/script-lc.xml", fx.Sample),
		fx.File("Basic/Standard/Module1.xml", fx.Sample)))
	// A script in another language, in a spreadsheet package.
	fx.Raw(t, "testdata", "scripts.ods", fx.Zip(
		mimetype("application/vnd.oasis.opendocument.spreadsheet"),
		fx.File("content.xml", fx.Sample),
		fx.File("Scripts/python/sample.py", fx.Sample)))
	// An empty Basic/ directory entry holds no macro.
	fx.Raw(t, "testdata", "empty-basic-dir.odt", odt(fx.File("Basic/", nil)))
	// A Basic directory that is not at the package root.
	fx.Raw(t, "testdata", "nested-basic.odt", odt(fx.File("Pictures/Basic/sample.png", fx.Sample)))
}
