// Command pdf.has-mixed-orientation answers: does the PDF mix portrait and landscape pages? A page's
// orientation is the one it is displayed in: its visible box turned by its
// /Rotate.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		if ctx.PageCount == 0 {
			return false, pdfx.ErrNoPages
		}
		var portrait, landscape bool
		for p := 1; p <= ctx.PageCount; p++ {
			box, err := pdfx.PageBox(ctx, p)
			if err != nil {
				return false, err
			}
			r, err := pdfx.PageRotate(ctx, p)
			if err != nil {
				return false, err
			}
			if r, err = pdfx.NormRotate(r); err != nil {
				return false, err
			}
			pt, ls := pdfx.Orientation(box, r)
			portrait, landscape = portrait || pt, landscape || ls
			if portrait && landscape {
				return true, nil
			}
		}
		return false, nil
	})
}
