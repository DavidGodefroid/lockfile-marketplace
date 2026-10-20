// Command pdf.has-outline answers: does the PDF have an outline (bookmarks)? Only the presence
// of a first item is tested; no title is read.
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
		outlines, err := pdfx.CatalogDict(ctx, "Outlines")
		if err != nil || outlines == nil {
			return false, err
		}
		first, found := outlines.Find("First")
		return found && first != nil, nil
	})
}
