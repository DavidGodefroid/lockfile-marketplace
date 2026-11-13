// Answers whether the file contains the typed needle, and writes the
// offset of each match as its detail. It fails when the needle reaches it
// anywhere but fd 6 (argv, env).
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		in, err := sdk.Inputs()
		if err != nil {
			return false, err
		}
		needle := in["needle"]
		if needle == "" {
			return false, errors.New("no needle")
		}
		for _, s := range append(os.Args, os.Environ()...) {
			if strings.Contains(s, needle) {
				return false, errors.New("the needle is not on fd 6 only")
			}
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return false, err
		}
		found := false
		for off := 0; ; {
			i := bytes.Index(b[off:], []byte(needle))
			if i < 0 {
				break
			}
			found = true
			fmt.Fprintf(sdk.Detail(), "found at offset %d\n", off+i)
			off += i + 1
		}
		return found, nil
	}, sdk.WithDetail())
}
