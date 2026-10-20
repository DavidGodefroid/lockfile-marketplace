// Command zip.uses-zip64 answers: does the ZIP archive use the ZIP64
// extensions?
package main

import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

const (
	eocdSig    = 0x06054b50 // end of central directory record
	locatorSig = 0x07064b50 // zip64 end of central directory locator
	eocdLen    = 22
	locatorLen = 20
	maxComment = 0xffff
	zip64Extra = 0x0001 // zip64 extended information extra field
)

var errNoEOCD = errors.New("end of central directory record not found")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		for _, e := range r.File {
			if hasZip64Extra(e.Extra) {
				return true, nil
			}
		}
		return hasLocator(f)
	})
}

// hasZip64Extra reports whether a central directory extra field holds a
// zip64 extended information block. A truncated trailing field ends the
// scan, as in archive/zip.
func hasZip64Extra(extra []byte) bool {
	for len(extra) >= 4 {
		tag := binary.LittleEndian.Uint16(extra)
		size := int(binary.LittleEndian.Uint16(extra[2:]))
		extra = extra[4:]
		if size > len(extra) {
			return false
		}
		if tag == zip64Extra {
			return true
		}
		extra = extra[size:]
	}
	return false
}

// hasLocator reports whether a zip64 end of central directory locator
// immediately precedes the end of central directory record.
func hasLocator(f *os.File) (bool, error) {
	st, err := f.Stat()
	if err != nil {
		return false, err
	}
	n := min(st.Size(), int64(locatorLen+eocdLen+maxComment))
	tail := make([]byte, n)
	if _, err := f.ReadAt(tail, st.Size()-n); err != nil {
		return false, err
	}
	i := eocdIndex(tail)
	if i < 0 {
		return false, errNoEOCD
	}
	return i >= locatorLen && binary.LittleEndian.Uint32(tail[i-locatorLen:]) == locatorSig, nil
}

// eocdIndex returns the offset in b of the last end of central directory
// record whose comment fits in b, the record archive/zip reads, or -1.
func eocdIndex(b []byte) int {
	for i := len(b) - eocdLen; i >= 0; i-- {
		if binary.LittleEndian.Uint32(b[i:]) != eocdSig {
			continue
		}
		if i+eocdLen+int(binary.LittleEndian.Uint16(b[i+20:])) <= len(b) {
			return i
		}
	}
	return -1
}
