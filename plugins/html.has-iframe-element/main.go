// Command html.has-iframe-element answers: does the file contain an HTML
// iframe start tag?
package main

import (
	"bufio"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasIframe(f)
	})
}

const tag = "<iframe"

// hasIframe reports whether r holds a case-insensitive "<iframe" followed by
// HTML whitespace, '/' or '>'. It scans byte by byte with a match counter, so
// a tag split across read-buffer boundaries is found. Read errors other than
// io.EOF are returned.
func hasIframe(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	matched := 0
	for {
		b, err := br.ReadByte()
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if matched == len(tag) {
			switch b {
			case ' ', '\t', '\n', '\f', '\r', '/', '>':
				return true, nil
			}
			matched = 0
		}
		// '<' occurs only at the start of tag, so a mismatch restarts the
		// match at 0, or at 1 when the mismatching byte is itself '<'.
		switch {
		case lower(b) == tag[matched]:
			matched++
		case b == '<':
			matched = 1
		default:
			matched = 0
		}
	}
}

func lower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + 'a' - 'A'
	}
	return b
}
