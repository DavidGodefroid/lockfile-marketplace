package main

import (
	"archive/zip"
	"bytes"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	creatorFAT  = 0
	creatorUnix = 3
	unixSymlink = 0o120000 // S_IFLNK
)

// moded is an entry with explicit creator and external attributes.
type moded struct {
	name    string
	body    []byte
	creator uint16
	attrs   uint32
}

// zipModed builds an archive whose entries carry raw mode bits, which
// formatxtest.Entry cannot express.
func zipModed(entries ...moded) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Store}
		h.CreatorVersion = e.creator<<8 | 20
		h.ExternalAttrs = e.attrs
		fw, err := w.CreateHeader(h)
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
	target := []byte("sample.txt")
	symlink := moded{"link", target, creatorUnix, (unixSymlink | 0o777) << 16}
	regular := moded{"sample.txt", body, creatorUnix, 0o100644 << 16}
	dir := moded{"dir/", nil, creatorUnix, 0o040755<<16 | 0x10}
	// Symlink type bits under an MS-DOS creator are not a Unix mode.
	fatHigh := moded{"fat.txt", body, creatorFAT, (unixSymlink | 0o777) << 16}
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "plain.zip", fx.Zip(fx.File("a.txt", body)))
	fx.Raw(t, "testdata", "unix-regular.zip", zipModed(dir, regular))
	fx.Raw(t, "testdata", "fat-symlink-bits.zip", zipModed(fatHigh))
	fx.Raw(t, "testdata", "symlink.zip", zipModed(regular, symlink))
	fx.Raw(t, "testdata", "symlink-only.zip", zipModed(symlink))
}
