// Command json.parses answers: does the file parse as exactly one JSON
// value?
package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/utf8x"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return parses(f)
	})
}

// parses reports whether r holds exactly one JSON value. It streams tokens
// rather than decoding the whole value, so memory stays bounded by the
// largest single string or number. A syntax error, a truncated value or
// invalid UTF-8 is a verdict; any other read error is returned.
func parses(r io.Reader) (bool, error) {
	dec := json.NewDecoder(utf8x.NewReader(r))
	depth, values := 0, 0
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return depth == 0 && values == 1, nil
		}
		var serr *json.SyntaxError
		if errors.As(err, &serr) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, utf8x.ErrInvalid) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if d, ok := tok.(json.Delim); ok {
			if d == '[' || d == '{' {
				depth++
				continue
			}
			depth--
		}
		if depth == 0 {
			values++
			if values > 1 {
				return false, nil
			}
		}
	}
}
