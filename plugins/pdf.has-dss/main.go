// Command pdf.has-dss answers: does the PDF carry a Document Security Store (long-term
// validation data)? Only the presence of the catalog /DSS dictionary is read; its content
// is not checked.
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
		dss, err := pdfx.CatalogDict(ctx, "DSS")
		return dss != nil, err
	})
}
