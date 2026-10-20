package main

import (
	"archive/zip"
	"hash/crc32"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	body := []byte("sample body\n")
	sum := crc32.ChecksumIEEE(body)
	valid := fx.Zip(
		fx.File("a.txt", body),
		fx.Entry{Name: "b.txt", Body: body, Method: zip.Store},
	)
	fx.Raw(t, "testdata", "empty.bin", nil)
	fx.Raw(t, "testdata", "sample.txt", fx.Sample)
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "valid.zip", valid)
	fx.Raw(t, "testdata", "sample.docx", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "insecure-name.zip", fx.Zip(fx.File("../outside.txt", body)))
	fx.Raw(t, "testdata", "truncated.zip", valid[:len(valid)/2])
	fx.Raw(t, "testdata", "bad-crc.zip", fx.Zip(
		fx.Entry{Name: "a.txt", Body: body, Method: zip.Store, Raw: true, CRC32: sum ^ 1},
	))
	fx.Raw(t, "testdata", "bad-deflate.zip", fx.Zip(
		fx.Entry{Name: "a.txt", Body: []byte{0xff, 0xff, 0xff, 0xff}, Method: zip.Deflate, Raw: true, CRC32: sum},
	))
	fx.Raw(t, "testdata", "encrypted-flag.zip", fx.Zip(
		fx.Entry{Name: "a.txt", Body: body, Method: zip.Store, Flags: 1, Raw: true, CRC32: sum},
	))
	fx.Raw(t, "testdata", "aes-method.zip", fx.Zip(
		fx.Entry{Name: "a.txt", Body: body, Method: 99, Raw: true, CRC32: sum},
	))
}
