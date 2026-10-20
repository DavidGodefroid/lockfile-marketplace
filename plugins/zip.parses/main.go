// Command zip.parses answers: does the file read as a complete, unencrypted
// ZIP archive?
package main

import (
	"archive/zip"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// methodAES is the WinZip AE-x method id; such entries are encrypted even
// without the traditional encryption flag.
const methodAES = 99

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, nil
		}
		for _, e := range r.File {
			if !readable(e) {
				return false, nil
			}
		}
		return true, nil
	})
}

// readable reports whether e is unencrypted and decompresses to the size
// and CRC-32 its header claims.
func readable(e *zip.File) bool {
	if e.Flags&1 != 0 || e.Method == methodAES {
		return false
	}
	return zipx.Drain(e) == nil
}
