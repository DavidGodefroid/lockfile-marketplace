// Command pdf.has-hidden-layer answers: does the PDF have a layer that is hidden by
// default? Only the /OFF array of the default optional-content configuration is read;
// the layers themselves are not.
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
		d, err := ctx.DereferenceDict(oc["D"])
		if err != nil || d == nil {
			return false, err
		}
		off, err := ctx.DereferenceArray(d["OFF"])
		if err != nil {
			return false, err
		}
		return len(off) > 0, nil
	})
}
