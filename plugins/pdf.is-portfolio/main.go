// Command pdf.is-portfolio answers: is the PDF a portfolio? Only the presence of the catalog
// /Collection dictionary is checked; the embedded files are not read.
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
		collection, err := pdfx.CatalogDict(ctx, "Collection")
		return collection != nil, err
	})
}
