// Command pdf.has-text-layer answers: does some page of the PDF show text,
// directly or through a form XObject, visible or not?
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
			dr, err := pdfx.PageDraws(ctx, p)
			if err != nil {
				return false, err
			}
			if dr.Text {
				return true, nil
			}
		}
		return false, nil
	})
}
