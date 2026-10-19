package main

import (
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		b := make([]byte, 1)
		n, err := f.Read(b)
		if err != nil && err != io.EOF {
			return false, err
		}
		return n == 1 && b[0] == 'Y', nil
	})
}
