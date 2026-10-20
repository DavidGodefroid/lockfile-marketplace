// Command text.valid-utf8 answers: is the file valid UTF-8?
package main

import (
	"bufio"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/utf8x"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return valid(f)
	})
}

// valid reports whether r is valid UTF-8. It streams r through a 64 KiB
// buffer, so memory does not grow with the file. An invalid sequence is a
// verdict; a read error is returned.
func valid(r io.Reader) (bool, error) {
	return utf8x.Valid(bufio.NewReaderSize(r, 64<<10))
}
