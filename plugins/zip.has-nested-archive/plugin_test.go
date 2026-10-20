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
	body := []byte("sample body\n")
	text := fx.File("a.txt", body)
	// Encrypted bodies are opaque bytes, written raw and never decrypted.
	zipCrypto := fx.Entry{Name: "secret.zip", Body: body, Method: zip.Store, Flags: 0x1, Raw: true}
	aesNoFlag := fx.Entry{Name: "secret.zip", Body: body, Method: 99, Raw: true}
	badDeflate := fx.Entry{Name: "b.bin", Body: []byte{0xff, 0xff, 0xff, 0xff}, Method: zip.Deflate, Raw: true}
	badCRC := fx.Entry{Name: "b.txt", Body: body, Method: zip.Store, Raw: true, CRC32: 1}
	unknownMethod := fx.Entry{Name: "b.bin", Body: body, Method: 12, Raw: true}
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt":     fx.Sample,
		"empty":          nil,
		"truncated":      fx.Zip(fx.File("a.txt", []byte("x")))[:20],
		"zipcrypto":      fx.Zip(text, zipCrypto),
		"aes-no-flag":    fx.Zip(text, aesNoFlag),
		"bad-deflate":    fx.Zip(text, badDeflate),
		"bad-crc":        fx.Zip(text, badCRC),
		"unknown-method": fx.Zip(text, unknownMethod),
	})
}
