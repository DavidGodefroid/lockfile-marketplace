// Command pdf.has-mixed-page-sizes answers: does the PDF have pages of different sizes?
// Orientation is ignored: A4 portrait and A4 landscape are the same size.
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
		first, err := pdfx.PageBox(ctx, 1)
		if err != nil {
			return false, err
		}
		for p := 2; p <= ctx.PageCount; p++ {
			box, err := pdfx.PageBox(ctx, p)
			if err != nil {
				return false, err
			}
			if !pdfx.SameSize(first, box) {
				return true, nil
			}
		}
		return false, nil
	})
}
