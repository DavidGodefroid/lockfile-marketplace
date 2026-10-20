package main

import (
	"archive/zip"
	"bytes"
	"hash/crc32"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// sized is a stored entry whose header declares the given uncompressed size
// over a compressed body of its own length, which formatxtest.Entry cannot
// express.
type sized struct {
	name string
	body []byte
	size uint64
}

func zipSized(entries ...sized) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Store}
		h.CRC32 = crc32.ChecksumIEEE(e.body)
		h.CompressedSize64 = uint64(len(e.body))
		h.UncompressedSize64 = e.size
		fw, err := w.CreateRaw(h)
		if err != nil {
			panic(err)
		}
		fw.Write(e.body)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	ten := bytes.Repeat([]byte("s"), 10)
	zeros := make([]byte, 1<<20)
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "plain.zip", fx.Zip(fx.File("a.txt", body), fx.File("dir/", nil)))
	fx.Raw(t, "testdata", "ratio-100.zip", zipSized(sized{"a", ten, 1000}))
	fx.Raw(t, "testdata", "ratio-over-100.zip", zipSized(sized{"a", ten, 1001}))
	fx.Raw(t, "testdata", "zero-compressed.zip", zipSized(sized{"a", nil, 1}))
	fx.Raw(t, "testdata", "deflated-zeros.zip", fx.Zip(fx.File("a.txt", body), fx.File("zeros", zeros)))
}
