// Command zip.has-executable-entry answers: does the ZIP archive contain an
// entry that is a native executable?
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// prefixLen covers the longest check: a fat Mach-O magic and its arch count.
const prefixLen = 8

// methodAES is the WinZip AES method, whose entries may lack flag bit 0.
const methodAES = 99

// maxFatArch bounds the arch count of a fat Mach-O header. Java class files
// share the 0xcafebabe magic but carry their version (≥ 45) there instead.
const maxFatArch = 20

var errEncrypted = errors.New("encrypted entry")

// signatures are the leading magic bytes of the executable formats the
// question covers.
var signatures = [][]byte{
	[]byte("MZ"),             // DOS / Windows PE
	[]byte("\x7fELF"),        // ELF
	{0xfe, 0xed, 0xfa, 0xce}, // Mach-O 32-bit, big endian
	{0xce, 0xfa, 0xed, 0xfe}, // Mach-O 32-bit, little endian
	{0xfe, 0xed, 0xfa, 0xcf}, // Mach-O 64-bit, big endian
	{0xcf, 0xfa, 0xed, 0xfe}, // Mach-O 64-bit, little endian
}

// fatMagics are the universal (fat) Mach-O magics, always big endian.
var fatMagics = [][]byte{
	{0xca, 0xfe, 0xba, 0xbe}, // fat
	{0xca, 0xfe, 0xba, 0xbf}, // fat 64
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
			if isExecutable(p) {
				return true, nil
			}
			// A 0 covers whole entries: a body that fails its CRC-32 or
			// decompression is a read failure, not "no executable".
			if err := zipx.Drain(e); err != nil {
				return false, err
			}
		}
		return false, nil
	})
}

// isExecutable reports whether p, the start of an entry, carries one of the
// signatures, or a fat Mach-O magic followed by a plausible arch count.
func isExecutable(p []byte) bool {
	for _, s := range signatures {
		if bytes.HasPrefix(p, s) {
			return true
		}
	}
	if len(p) < 8 {
		return false
	}
	for _, s := range fatMagics {
		if bytes.HasPrefix(p, s) {
			n := binary.BigEndian.Uint32(p[4:8])
			return n > 0 && n < maxFatArch
		}
	}
	return false
}
