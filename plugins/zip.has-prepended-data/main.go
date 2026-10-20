// Command zip.has-prepended-data answers: does the ZIP archive have data
// before its first record, as a self-extractor or a polyglot does?
package main

import (
	"bytes"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

var (
	localHeader = []byte("PK\x03\x04")
	endOfDir    = []byte("PK\x05\x06")
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		head := make([]byte, len(localHeader))
		if _, err := f.ReadAt(head, 0); err != nil && err != io.EOF {
			return false, err
		}
		if bytes.Equal(head, localHeader) {
			return false, nil
		}
		// An archive without entries is only its end of central directory.
		if len(r.File) == 0 && bytes.Equal(head, endOfDir) {
			return false, nil
		}
		return true, nil
	})
}
