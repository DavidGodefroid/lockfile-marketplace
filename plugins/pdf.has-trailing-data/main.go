// Command pdf.has-trailing-data answers: does the PDF have data after its last %%EOF?
package main

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// chunk is how much of the file one backward step reads.
const chunk = 64 << 10

var (
	eofMarker = []byte("%%EOF")
	errNoEOF  = errors.New("no %%EOF in the file")
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		if _, err := pdfx.Open(f); err != nil {
			return false, err
		}
		fi, err := f.Stat()
		if err != nil {
			return false, err
		}
		return trailingData(f, fi.Size())
	})
}

// trailingData scans r backwards from its end, chunk by chunk, to the last
// %%EOF and reports whether a byte other than PDF whitespace follows it. A
// file without %%EOF is an error.
func trailingData(r io.ReaderAt, size int64) (bool, error) {
	junk := int64(-1) // offset of the last non-whitespace byte seen so far
	for end := size; end > 0; {
		start := max(end-chunk, 0)
		// Overlap the previous window by len(marker)-1 bytes so a marker
		// split across two windows is still found.
		hi := min(end+int64(len(eofMarker))-1, size)
		buf := make([]byte, hi-start)
		if _, err := r.ReadAt(buf, start); err != nil && !errors.Is(err, io.EOF) {
			return false, err
		}
		if junk < 0 {
			if j := lastNonBlank(buf); j >= 0 {
				junk = start + int64(j)
			}
		}
		if i := bytes.LastIndex(buf, eofMarker); i >= 0 {
			return junk >= start+int64(i+len(eofMarker)), nil
		}
		end = start
	}
	return false, errNoEOF
}

// lastNonBlank returns the index of the last byte of b that is not PDF
// whitespace (ISO 32000-1, 7.2.2), or -1.
func lastNonBlank(b []byte) int {
	for i := len(b) - 1; i >= 0; i-- {
		switch b[i] {
		case 0, '\t', '\n', '\f', '\r', ' ':
		default:
			return i
		}
	}
	return -1
}
