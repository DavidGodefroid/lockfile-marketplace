// Command pdf.has-non-a4-page answers: does the PDF have at least one page
// whose visible size is not A4?
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
			box, err := pdfx.PageBox(ctx, p)
			if err != nil {
				return false, err
			}
			if !pdfx.IsA4(box.Size()) {
				return true, nil
			}
		}
		return false, nil
	})
}
