// Command csv.consistent-columns answers: does the file parse as
// comma-separated values with the same number of columns on every row?
package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/utf8x"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return consistent(f)
	})
}

// consistent reports whether r holds at least one CSV record and every
// record parses with as many fields as the first. A parse error or invalid
// UTF-8 is a verdict; any other read error is returned.
func consistent(r io.Reader) (bool, error) {
	c := csv.NewReader(utf8x.NewReader(bufio.NewReader(r)))
	c.ReuseRecord = true
	records := 0
	for {
		_, err := c.Read()
		if errors.Is(err, io.EOF) {
			return records > 0, nil
		}
		var perr *csv.ParseError
		if errors.Is(err, utf8x.ErrInvalid) || errors.As(err, &perr) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		records++
	}
}
