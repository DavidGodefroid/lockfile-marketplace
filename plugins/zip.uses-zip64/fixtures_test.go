package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	classic := fx.Zip(fx.File("a.txt", body), fx.Entry{Name: "b.txt", Body: body, Method: zip.Store})
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "classic.zip", classic)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "other-extra.zip", withModified(body))
	fx.Raw(t, "testdata", "zip64-size.zip", withZip64Size(body))
	fx.Raw(t, "testdata", "zip64-end.zip", withZip64End(classic, true))
	fx.Raw(t, "testdata", "zip64-end-unneeded.zip", withZip64End(classic, false))
}

// withModified is a classic archive whose entry carries an extended
// timestamp extra field (0x5455), not a zip64 one.
func withModified(body []byte) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, err := w.CreateHeader(&zip.FileHeader{Name: "a.txt", Method: zip.Store, Modified: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		panic(err)
	}
	fw.Write(body)
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// withZip64Size is an archive whose raw entry claims a 4 GiB uncompressed
// size, so archive/zip writes a zip64 extra field in its central directory
// header. The body is never read.
func withZip64Size(body []byte) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fw, err := w.CreateRaw(&zip.FileHeader{
		Name:               "big.bin",
		Method:             zip.Deflate,
		CompressedSize64:   uint64(len(body)),
		UncompressedSize64: 1 << 32,
	})
	if err != nil {
		panic(err)
	}
	fw.Write(body)
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// withZip64End inserts a zip64 end of central directory record and its
// locator before the end of central directory record of z (which must have
// no comment). With needed, the classic record's offset is set to
// 0xFFFFFFFF so readers must follow the locator.
func withZip64End(z []byte, needed bool) []byte {
	le := binary.LittleEndian
	end := len(z) - eocdLen
	eocd := append([]byte{}, z[end:]...)
	entries := uint64(le.Uint16(eocd[10:]))
	cdSize := uint64(le.Uint32(eocd[12:]))
	cdOffset := uint64(le.Uint32(eocd[16:]))

	out := append([]byte{}, z[:end]...)
	out = le.AppendUint32(out, 0x06064b50)
	out = le.AppendUint64(out, 44)
	out = le.AppendUint16(out, 45)
	out = le.AppendUint16(out, 45)
	out = le.AppendUint32(out, 0)
	out = le.AppendUint32(out, 0)
	out = le.AppendUint64(out, entries)
	out = le.AppendUint64(out, entries)
	out = le.AppendUint64(out, cdSize)
	out = le.AppendUint64(out, cdOffset)

	out = le.AppendUint32(out, locatorSig)
	out = le.AppendUint32(out, 0)
	out = le.AppendUint64(out, uint64(end))
	out = le.AppendUint32(out, 1)

	if needed {
		le.PutUint32(eocd[16:], 0xffffffff)
	}
	return append(out, eocd...)
}
