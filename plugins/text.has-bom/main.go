// Command text.has-bom answers: does the file start with a Unicode byte
// order mark?
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
		return hasBOM(f)
	})
}

// boms are the recognised byte order marks, written as escaped bytes so the
// source holds no literal BOM. FF FE also covers UTF-32 LE (FF FE 00 00).
var boms = [][]byte{
	{0xEF, 0xBB, 0xBF},       // UTF-8
	{0xFE, 0xFF},             // UTF-16 BE
	{0xFF, 0xFE},             // UTF-16 LE, UTF-32 LE
	{0x00, 0x00, 0xFE, 0xFF}, // UTF-32 BE
}

// hasBOM reports whether r starts with one of boms. It reads at most four
// bytes; a file shorter than a mark answers false. Read errors other than
// io.EOF and io.ErrUnexpectedEOF are returned.
func hasBOM(r io.Reader) (bool, error) {
	head := make([]byte, 4)
	n, err := io.ReadFull(r, head)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false, err
	}
	for _, bom := range boms {
		if bytes.HasPrefix(head[:n], bom) {
			return true, nil
		}
	}
	return false, nil
}
