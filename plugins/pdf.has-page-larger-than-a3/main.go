// Command pdf.has-page-larger-than-a3 answers: does the PDF have at least
// one page whose visible size exceeds A3?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// A3 in points (297 x 420 mm), plus the 1 mm tolerance.
const (
	a3Short = pdfx.A4Long
	a3Long  = 1190.551
)

// largerThanA3 reports whether w x h does not fit A3 within 1 mm on each
// side, either orientation.
func largerThanA3(w, h float64) bool {
	short, long := min(w, h), max(w, h)
	return short > a3Short+pdfx.A4Tolerance || long > a3Long+pdfx.A4Tolerance
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		if ctx.PageCount == 0 {
			return false, pdfx.ErrNoPages
		}
		for p := 1; p <= ctx.PageCount; p++ {
			box, err := pdfx.PageBox(ctx, p)
			if err != nil {
				return false, err
			}
			if largerThanA3(box.Size()) {
				return true, nil
			}
		}
		return false, nil
	})
}
