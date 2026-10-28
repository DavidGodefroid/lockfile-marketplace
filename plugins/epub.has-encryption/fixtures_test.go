package main

import (
	"archive/zip"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const container = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`

const encryption = `<?xml version="1.0" encoding="UTF-8"?>
<encryption xmlns="urn:oasis:names:tc:opendocument:xmlns:container"/>`

// epub builds a minimal EPUB: stored mimetype first, then the container,
// a sample chapter and extra.
func epub(extra ...fx.Entry) []byte {
	entries := []fx.Entry{
		{Name: "mimetype", Body: []byte(mimetype), Method: zip.Store},
		fx.File("META-INF/container.xml", []byte(container)),
		fx.File("OEBPS/content.opf", []byte("<package/>")),
		fx.File("OEBPS/chapter.xhtml", []byte("<html><body><p>sample</p></body></html>")),
	}
	return fx.Zip(append(entries, extra...)...)
}

func writeFixtures(t *testing.T) {
	enc := fx.File("META-INF/encryption.xml", []byte(encryption))
	fx.Raw(t, "testdata", "plain.epub", epub())
	fx.Raw(t, "testdata", "encryption-elsewhere.epub", epub(fx.File("OEBPS/encryption.xml", []byte(encryption))))
	fx.Raw(t, "testdata", "encryption-wrong-case.epub", epub(fx.File("meta-inf/encryption.xml", []byte(encryption))))
	fx.Raw(t, "testdata", "encryption-dir.epub", epub(fx.Entry{Name: "META-INF/encryption.xml/", Method: zip.Store}))
	fx.Raw(t, "testdata", "signatures-only.epub", epub(fx.File("META-INF/signatures.xml", []byte("<signatures/>"))))
	fx.Raw(t, "testdata", "encrypted.epub", epub(enc))
	fx.Raw(t, "testdata", "encrypted-stored.epub", epub(fx.Entry{Name: "META-INF/encryption.xml", Body: []byte(encryption), Method: zip.Store}))
	fx.Raw(t, "testdata", "encrypted-empty.epub", epub(fx.Entry{Name: "META-INF/encryption.xml", Method: zip.Store}))
}
