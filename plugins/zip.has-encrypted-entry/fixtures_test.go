package main

import (
	"archive/zip"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	plain := fx.File("a.txt", body)
	// Encrypted bodies are opaque bytes: the answer reads only the central
	// directory, so they are written raw and never decrypted.
	zipCrypto := fx.Entry{Name: "secret.txt", Body: body, Method: zip.Store, Flags: 0x1, Raw: true}
	zipCryptoDeflate := fx.Entry{Name: "secret.txt", Body: body, Method: zip.Deflate, Flags: 0x1, Raw: true}
	aes := fx.Entry{Name: "secret.txt", Body: body, Method: 99, Flags: 0x1, Raw: true}
	aesNoFlag := fx.Entry{Name: "secret.txt", Body: body, Method: 99, Raw: true}
	otherFlags := fx.Entry{Name: "b.txt", Body: body, Method: zip.Deflate, Flags: 0x800 | 0x2}
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "plain.zip", fx.Zip(plain, fx.Entry{Name: "c.txt", Body: body, Method: zip.Store}))
	fx.Raw(t, "testdata", "other-flags.zip", fx.Zip(plain, otherFlags))
	fx.Raw(t, "testdata", "zipcrypto-store.zip", fx.Zip(plain, zipCrypto))
	fx.Raw(t, "testdata", "zipcrypto-deflate.zip", fx.Zip(zipCryptoDeflate))
	fx.Raw(t, "testdata", "aes.zip", fx.Zip(plain, aes))
	fx.Raw(t, "testdata", "aes-no-flag.zip", fx.Zip(aesNoFlag))
}
