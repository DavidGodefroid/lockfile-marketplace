// Command pdf.declares-language answers: does the PDF declare a document language? The /Lang value
// is never returned or kept: only whether it is empty.
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
		return pdfx.CatalogNonEmptyString(ctx, "Lang")
	})
}
