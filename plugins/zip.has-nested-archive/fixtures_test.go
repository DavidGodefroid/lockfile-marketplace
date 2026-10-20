package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	text := fx.File("a.txt", []byte("sample body\n"))
	dir := fx.Entry{Name: "dir/", Method: zip.Store}
	with := func(body []byte) []byte {
		return fx.Zip(dir, text, fx.File("dir/nested.bin", body))
	}
	// Entries that come close to a signature without carrying one.
	ustarAt256 := append(bytes.Repeat([]byte{' '}, 256), "ustar"...)
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "plain.zip", fx.Zip(dir, text, fx.Entry{Name: "b.txt", Body: []byte("stored\n"), Method: zip.Store}))
	fx.Raw(t, "testdata", "dirs-only.zip", fx.Zip(dir, fx.Entry{Name: "dir/sub/", Method: zip.Store}))
	fx.Raw(t, "testdata", "near-misses.zip", fx.Zip(
		fx.File("short.bin", []byte("PK")),
		fx.File("pk-other.bin", []byte("PK\x01\x02 central directory")),
		fx.File("bzh-letter.txt", []byte("BZhx not bzip2")),
		fx.File("ustar-at-0.txt", []byte("ustar at the start")),
		fx.File("ustar-at-256.txt", ustarAt256),
		fx.File("gzip-one-byte.bin", []byte{0x1f}),
		fx.File("rar-short.bin", []byte("Rar!\x1a")),
	))
	fx.Raw(t, "testdata", "docx-as-zip.zip", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "zip.zip", with(fx.Zip(text)))
	fx.Raw(t, "testdata", "zip-stored.zip", fx.Zip(fx.Entry{Name: "inner.zip", Body: fx.Zip(text), Method: zip.Store}))
	fx.Raw(t, "testdata", "zip-empty.zip", with(fx.Zip()))
	fx.Raw(t, "testdata", "zip-spanned.zip", with([]byte("PK\x07\x08 spanned")))
	fx.Raw(t, "testdata", "gzip.zip", with(gz(t)))
	fx.Raw(t, "testdata", "bzip2.zip", with([]byte("BZh91AY&SY")))
	fx.Raw(t, "testdata", "xz.zip", with([]byte("\xfd7zXZ\x00\x00\x04")))
	fx.Raw(t, "testdata", "zstd.zip", with([]byte("\x28\xb5\x2f\xfd\x00")))
	fx.Raw(t, "testdata", "7z.zip", with([]byte("7z\xbc\xaf\x27\x1c\x00\x04")))
	fx.Raw(t, "testdata", "rar.zip", with([]byte("Rar!\x1a\x07\x01\x00")))
	fx.Raw(t, "testdata", "tar.zip", with(tarball(t)))
	fx.Raw(t, "testdata", "nested-after-dir.zip", fx.Zip(fx.Entry{Name: "x.zip/", Method: zip.Store}, fx.File("x.zip/inner.gz", gz(t))))
}

func gz(t *testing.T) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte("sample body\n")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func tarball(t *testing.T) []byte {
	var b bytes.Buffer
	w := tar.NewWriter(&b)
	body := []byte("sample body\n")
	if err := w.WriteHeader(&tar.Header{Name: "a.txt", Mode: 0o644, Size: int64(len(body)), Format: tar.FormatUSTAR}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
