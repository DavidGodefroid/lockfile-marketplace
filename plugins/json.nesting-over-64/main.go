// Command json.nesting-over-64 answers: does the JSON value nest arrays and
// objects more than 64 levels deep?
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/utf8x"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxDepth is the deepest array/object nesting answered with 0.
const maxDepth = 64

// errInvalid marks input that is not exactly one well-formed JSON value.
var errInvalid = errors.New("not a single valid JSON value")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return nestingOver(f, maxDepth)
	})
}

// nestingOver reports whether the single JSON value in r nests arrays and
// objects more than limit levels deep. It streams tokens, so memory stays
// bounded by the nesting depth and the largest single string or number. The
// whole input is validated with the rules of json.parses: a syntax error, a
// truncated value, invalid UTF-8, a BOM, no value or data after the first
// value is an error, never a verdict.
func nestingOver(r io.Reader, limit int) (bool, error) {
	dec := json.NewDecoder(utf8x.NewReader(r))
	depth, deepest, values := 0, 0, 0
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			if depth != 0 || values != 1 {
				return false, errInvalid
			}
			return deepest > limit, nil
		}
		var serr *json.SyntaxError
		if errors.As(err, &serr) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, utf8x.ErrInvalid) {
			return false, fmt.Errorf("%w: %w", errInvalid, err)
		}
		if err != nil {
			return false, err
		}
		if d, ok := tok.(json.Delim); ok {
			if d == '[' || d == '{' {
				depth++
				deepest = max(deepest, depth)
				continue
			}
			depth--
		}
		if depth == 0 {
			values++
			if values > 1 {
				return false, errInvalid
			}
		}
	}
}
