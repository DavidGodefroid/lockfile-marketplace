// Command zip.has-nested-archive answers: does the ZIP archive contain an
// entry that is itself an archive or a compressed stream?
package main

import (
	"bytes"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// prefixLen covers the tar magic at offset 257 with room to spare.
const prefixLen = 512

// methodAES is the WinZip AES method, whose entries may lack flag bit 0.
const methodAES = 99

var errEncrypted = errors.New("encrypted entry")

// signatures are the leading magic bytes of the archive and compressed
// stream formats the question covers.
var signatures = [][]byte{
	[]byte("PK\x03\x04"),                 // zip local file header
	[]byte("PK\x05\x06"),                 // zip end of central directory (empty zip)
	[]byte("PK\x07\x08"),                 // zip spanned marker
	{0x1f, 0x8b},                         // gzip
	{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}, // xz
	{0x28, 0xb5, 0x2f, 0xfd},             // zstd
	{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}, // 7z
	[]byte("Rar!\x1a\x07"),               // rar 1.5 to 5
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		for _, e := range r.File {
			if e.FileInfo().IsDir() {
				continue
			}
			if e.Flags&0x1 != 0 || e.Method == methodAES {
				return false, errEncrypted
			}
			p, err := zipx.Prefix(e, prefixLen)
			if err != nil {
				return false, err
			}
			if isArchive(p) {
				return true, nil
			}
		}
		return false, nil
	})
}

// isArchive reports whether p, the start of an entry, carries one of the
// signatures, "BZh" and a block-size digit (bzip2), or "ustar" at offset
// 257 (tar).
func isArchive(p []byte) bool {
	for _, s := range signatures {
		if bytes.HasPrefix(p, s) {
			return true
		}
	}
	if len(p) >= 4 && bytes.HasPrefix(p, []byte("BZh")) && '0' <= p[3] && p[3] <= '9' {
		return true
	}
	return len(p) >= 262 && bytes.Equal(p[257:262], []byte("ustar"))
}
