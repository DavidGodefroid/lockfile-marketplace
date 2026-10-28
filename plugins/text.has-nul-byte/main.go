// Command text.has-nul-byte answers: does the file contain a NUL byte?
package main

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasNul(f)
	})
}

// hasNul reports whether r holds a 0x00 byte. It reads in fixed-size chunks
// and stops at the first NUL. Read errors other than io.EOF are returned.
func hasNul(r io.Reader) (bool, error) {
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if bytes.IndexByte(buf[:n], 0) >= 0 {
			return true, nil
		}
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
}
