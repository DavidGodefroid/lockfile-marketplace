package main

import (
	"encoding/binary"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// withComment sets the archive comment of a zip built by fx.Zip: its
// end-of-central-directory record ends with a zero comment length.
func withComment(b []byte, comment string) []byte {
	out := append([]byte(nil), b...)
	binary.LittleEndian.PutUint16(out[len(out)-2:], uint16(len(comment)))
	return append(out, comment...)
}

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	plain := fx.Zip(fx.File("a.txt", body), fx.File("b.txt", body))
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "plain.zip", plain)
	fx.Raw(t, "testdata", "comment.zip", withComment(plain, "sample comment"))
	fx.Raw(t, "testdata", "one-byte-comment.zip", withComment(plain, " "))
	fx.Raw(t, "testdata", "empty-archive-comment.zip", withComment(fx.Zip(), "sample comment"))
}
