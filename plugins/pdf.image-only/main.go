// Command pdf.image-only answers: does every page of the PDF draw an image,
// directly or through a form XObject, while no page shows text?
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
			if dr.Text || !dr.Image {
				return false, nil
			}
		}
		return true, nil
	})
}
