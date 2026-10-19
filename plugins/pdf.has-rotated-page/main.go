// Command pdf.has-rotated-page answers: does the PDF have at least one page
// with a /Rotate other than 0?
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
		for p := 1; p <= ctx.PageCount; p++ {
			raw, err := pdfx.PageRotate(ctx, p)
			if err != nil {
				return false, err
			}
			r, err := pdfx.NormRotate(raw)
			if err != nil {
				return false, err
			}
			if r != 0 {
				return true, nil
			}
		}
		return false, nil
	})
}
