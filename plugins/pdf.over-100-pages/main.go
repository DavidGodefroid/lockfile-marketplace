// Command pdf.over-100-pages answers: does the PDF have more than 100 pages?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxPages is the largest page count answered 0.
const maxPages = 100

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return ctx.PageCount > maxPages, nil
	})
}
