// Command pdf.has-page-labels answers: does the PDF define page labels? Only the presence
// of the catalog /PageLabels number tree is checked; the labels are not read.
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
		labels, err := pdfx.CatalogDict(ctx, "PageLabels")
		return labels != nil, err
	})
}
