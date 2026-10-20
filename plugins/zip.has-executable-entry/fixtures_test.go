package main

import (
	"archive/zip"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	text := fx.File("a.txt", []byte("sample body\n"))
	dir := fx.Entry{Name: "dir/", Method: zip.Store}
	with := func(body []byte) []byte {
		return fx.Zip(dir, text, fx.File("dir/sample.bin", body))
	}
	fx.Raw(t, "testdata", "empty-archive.zip", fx.Zip())
	fx.Raw(t, "testdata", "plain.zip", fx.Zip(dir, text, fx.Entry{Name: "b.txt", Body: []byte("stored\n"), Method: zip.Store}))
	fx.Raw(t, "testdata", "dirs-only.zip", fx.Zip(dir, fx.Entry{Name: "dir/sub/", Method: zip.Store}))
	// Entries that come close to a signature without carrying one.
	fx.Raw(t, "testdata", "near-misses.zip", fx.Zip(
		fx.File("m.txt", []byte("M")),
		fx.File("mz-lower.txt", []byte("mz sample")),
		fx.File("elf-short.bin", []byte("\x7fEL")),
		fx.File("elf-no-del.txt", []byte("ELF sample")),
		fx.File("macho-short.bin", []byte{0xfe, 0xed, 0xfa}),
		fx.File("java.class", []byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, 0x41}),
		fx.File("fat-zero.bin", []byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, 0x00}),
		fx.File("fat-short.bin", []byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00}),
		fx.File("pe-later.txt", []byte("sample MZ\x7fELF")),
	))
	fx.Raw(t, "testdata", "docx-as-zip.zip", fx.Docx(fx.Paragraph))
	fx.Raw(t, "testdata", "pe.zip", with([]byte("MZ\x90\x00sample")))
	fx.Raw(t, "testdata", "pe-stored.zip", fx.Zip(fx.Entry{Name: "sample.exe", Body: []byte("MZsample"), Method: zip.Store}))
	fx.Raw(t, "testdata", "elf.zip", with([]byte("\x7fELF\x02\x01\x01\x00sample")))
	fx.Raw(t, "testdata", "macho32-be.zip", with([]byte{0xfe, 0xed, 0xfa, 0xce, 0x00}))
	fx.Raw(t, "testdata", "macho32-le.zip", with([]byte{0xce, 0xfa, 0xed, 0xfe, 0x07}))
	fx.Raw(t, "testdata", "macho64-be.zip", with([]byte{0xfe, 0xed, 0xfa, 0xcf, 0x01}))
	fx.Raw(t, "testdata", "macho64-le.zip", with([]byte{0xcf, 0xfa, 0xed, 0xfe, 0x07}))
	fx.Raw(t, "testdata", "macho-fat.zip", with([]byte{0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, 0x02}))
	fx.Raw(t, "testdata", "macho-fat64.zip", with([]byte{0xca, 0xfe, 0xba, 0xbf, 0x00, 0x00, 0x00, 0x01}))
	fx.Raw(t, "testdata", "exe-after-dir.zip", fx.Zip(fx.Entry{Name: "x.exe/", Method: zip.Store}, fx.File("x.exe/inner", []byte("\x7fELF"))))
}
