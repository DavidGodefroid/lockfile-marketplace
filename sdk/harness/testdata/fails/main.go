package main

import (
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// Answers 1 on a file starting with Y, fails on anything else.
func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		b := make([]byte, 1)
		if n, _ := f.Read(b); n == 1 && b[0] == 'Y' {
			return true, nil
		}
		return false, errors.New("not a Y file")
	})
}
