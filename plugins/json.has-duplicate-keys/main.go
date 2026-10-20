// Command json.has-duplicate-keys answers: does some JSON object in the
// file repeat a key?
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

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasDuplicateKeys(f)
	})
}

// errNotOneValue reports input that is not exactly one JSON value.
var errNotOneValue = errors.New("not exactly one JSON value")

// frame is one open array or object. For an object, key tells whether the
// next token is a key and seen holds the keys met so far.
type frame struct {
	obj  bool
	key  bool
	seen map[string]struct{}
}

// hasDuplicateKeys reports whether some object in r has the same key twice.
// Keys are compared after unescaping. It streams tokens and walks the whole
// input even after a duplicate is found, so any input that is not exactly
// one valid JSON value (empty, BOM, invalid UTF-8, syntax error, truncated
// value, data after the value) is an error, never a verdict.
func hasDuplicateKeys(r io.Reader) (bool, error) {
	dec := json.NewDecoder(utf8x.NewReader(r))
	var stack []*frame
	dup, values := false, 0
	// valueDone marks a complete value: in an object, a key comes next.
	valueDone := func() {
		if n := len(stack); n > 0 && stack[n-1].obj {
			stack[n-1].key = true
		}
	}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			if len(stack) != 0 || values != 1 {
				return false, errNotOneValue
			}
			return dup, nil
		}
		if err != nil {
			return false, fmt.Errorf("%w: %w", errNotOneValue, err)
		}
		if len(stack) == 0 {
			values++
			if values > 1 {
				return false, errNotOneValue
			}
		}
		if d, ok := tok.(json.Delim); ok {
			switch d {
			case '{', '[':
				f := &frame{obj: d == '{', key: d == '{'}
				if f.obj {
					f.seen = map[string]struct{}{}
				}
				stack = append(stack, f)
			default:
				stack = stack[:len(stack)-1]
				valueDone()
			}
			continue
		}
		if n := len(stack); n > 0 && stack[n-1].key {
			top := stack[n-1]
			k, _ := tok.(string)
			if _, ok := top.seen[k]; ok {
				dup = true
			}
			top.seen[k] = struct{}{}
			top.key = false
			continue
		}
		valueDone()
	}
}
