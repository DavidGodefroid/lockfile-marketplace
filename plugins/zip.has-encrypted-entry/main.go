// Command zip.has-encrypted-entry answers: does the ZIP archive have an
// encrypted entry?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

const (
	flagEncrypted = 0x1 // general purpose bit 0: traditional or strong encryption
	methodAES     = 99  // WinZip AE-1/AE-2
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		for _, e := range r.File {
			if e.Flags&flagEncrypted != 0 || e.Method == methodAES {
				return true, nil
			}
		}
		return false, nil
	})
}
