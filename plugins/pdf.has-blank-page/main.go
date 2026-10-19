// Command pdf.has-blank-page answers: does the PDF have at least one page
// that paints nothing visible?
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
			paints, err := pdfx.PagePaints(ctx, p)
			if err != nil {
				return false, err
			}
			if !paints {
				return true, nil
			}
		}
		return false, nil
	})
}
