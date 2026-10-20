// Command pdf.has-output-intent answers: does the PDF have an output intent? Only the catalog
// /OutputIntents array is read; the ICC profile is not checked.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		intents, err := pdfx.CatalogArray(ctx, "OutputIntents")
		if err != nil {
			return false, err
		}
		for _, o := range intents {
			o, err := ctx.Dereference(o)
			if err != nil {
				return false, err
			}
			if _, ok := o.(types.Dict); ok {
				return true, nil
			}
		}
		return false, nil
	})
}
