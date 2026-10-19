// Command file.non-empty answers: does the file contain at least one byte?
package main

import (
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		b := make([]byte, 1)
		n, err := f.Read(b)
		switch {
		case n == 1:
			return true, nil
		case errors.Is(err, io.EOF):
			return false, nil
		}
		return false, err
	})
}
