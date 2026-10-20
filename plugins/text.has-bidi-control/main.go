// Command text.has-bidi-control answers: does the UTF-8 text hold a
// bidirectional override or isolate control?
package main

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/utf8x"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasBidiControl(f)
	})
}

// isBidiControl reports whether c is an embedding, override or isolate
// control: U+202A–U+202E (LRE, RLE, PDF, LRO, RLO) or U+2066–U+2069 (LRI,
// RLI, FSI, PDI).
func isBidiControl(c rune) bool {
	return (c >= 0x202A && c <= 0x202E) || (c >= 0x2066 && c <= 0x2069)
}

// hasBidiControl reports whether r is valid UTF-8 holding a bidi control.
// It streams r through a 64 KiB buffer and reads it to the end even after a
// match, so invalid UTF-8 anywhere is an error, never a verdict.
func hasBidiControl(r io.Reader) (bool, error) {
	br := bufio.NewReaderSize(utf8x.NewReader(r), 64<<10)
	found := false
	for {
		c, _, err := br.ReadRune()
		if errors.Is(err, io.EOF) {
			return found, nil
		}
		if err != nil {
			return false, err
		}
		if isBidiControl(c) {
			found = true
		}
	}
}
