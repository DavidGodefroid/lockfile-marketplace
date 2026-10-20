// Command pdf.is-tagged answers: is the PDF a tagged PDF? Only the catalog entries that
// declare it are read; the structure tree itself is not checked.
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
		mark, err := pdfx.CatalogDict(ctx, "MarkInfo")
		if err != nil {
			return false, err
		}
		tree, err := pdfx.CatalogDict(ctx, "StructTreeRoot")
		if err != nil || tree == nil || mark == nil {
			return false, err
		}
		marked := mark.BooleanEntry("Marked")
		return marked != nil && *marked, nil
	})
}
