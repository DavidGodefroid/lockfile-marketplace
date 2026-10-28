// Command pdf.has-optional-content answers: does the PDF have optional content (layers)? Only
// the catalog /OCProperties /OCGs array is counted; no layer name is read.
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
		oc, err := pdfx.CatalogDict(ctx, "OCProperties")
		if err != nil || oc == nil {
			return false, err
		}
		ocgs, err := ctx.DereferenceArray(oc["OCGs"])
		if err != nil {
			return false, err
		}
		return len(ocgs) > 0, nil
	})
}
