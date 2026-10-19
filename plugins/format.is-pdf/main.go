// Command format.is-pdf answers: does the file start like a PDF?
package main

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// window is how far into the file %PDF- may appear.
const window = 1024

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		b := make([]byte, window)
		n, err := io.ReadFull(f, b)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
			return false, err
		}
		return bytes.Contains(b[:n], []byte("%PDF-")), nil
	})
}
