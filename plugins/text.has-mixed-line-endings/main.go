// Command text.has-mixed-line-endings answers: does the file mix line
// ending styles?
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
		return hasMixedLineEndings(f)
	})
}

// Line ending kinds, as bits of a set.
const (
	crlf = 1 << iota
	lf
	cr
)

// hasMixedLineEndings reports whether r holds at least two of CRLF, lone LF
// and lone CR. It reads byte by byte through a buffer, so a CR LF pair split
// across reads still counts as one CRLF, and stops as soon as two kinds are
// seen. Read errors other than io.EOF are returned.
func hasMixedLineEndings(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	seen := 0
	pendingCR := false
	for {
		b, err := br.ReadByte()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return false, err
			}
			if pendingCR {
				seen |= cr
			}
			return mixed(seen), nil
		}
		switch {
		case b == '\n' && pendingCR:
			seen |= crlf
			pendingCR = false
		case b == '\n':
			seen |= lf
		case pendingCR:
			seen |= cr
			pendingCR = b == '\r'
		default:
			pendingCR = b == '\r'
		}
		if mixed(seen) {
			return true, nil
		}
	}
}

// mixed reports whether set holds at least two kinds.
func mixed(set int) bool {
	return set&(set-1) != 0
}
