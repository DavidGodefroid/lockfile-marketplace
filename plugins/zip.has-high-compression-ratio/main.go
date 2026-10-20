// Command zip.has-high-compression-ratio answers: does the ZIP archive have
// an entry whose uncompressed size is more than 100 times its compressed
// size?
package main

import (
	"math"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxRatio is the highest uncompressed/compressed ratio still answered 0.
const maxRatio = 100

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		// The sizes are the ones the central directory declares; no entry is
		// decompressed, so a bomb costs nothing to detect.
		for _, e := range r.File {
			if highRatio(e.UncompressedSize64, e.CompressedSize64) {
				return true, nil
			}
		}
		return false, nil
	})
}

// highRatio reports u/c > maxRatio without dividing; c = 0 with u > 0 is an
// unbounded ratio.
func highRatio(u, c uint64) bool {
	if c > math.MaxUint64/maxRatio {
		return false
	}
	return u > c*maxRatio
}
