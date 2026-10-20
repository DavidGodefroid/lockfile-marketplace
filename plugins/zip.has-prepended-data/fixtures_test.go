package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	archive := fx.Zip(fx.File("a.txt", body), fx.File("b.txt", body))
	// archive/zip locates entries relative to where the central directory
	// really is, so bytes put in front still parse as the same archive.
	stub := []byte("MZ sample stub, not an executable\n")
	fx.Raw(t, "testdata", "plain.zip", archive)
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "stub-prefix.zip", append(append([]byte{}, stub...), archive...))
	fx.Raw(t, "testdata", "one-byte-prefix.zip", append([]byte{0}, archive...))
	fx.Raw(t, "testdata", "prefix-empty-archive.zip", append(append([]byte{}, stub...), fx.Zip()...))
}
