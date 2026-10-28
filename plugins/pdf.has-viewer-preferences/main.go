// Command pdf.has-viewer-preferences answers: does the PDF set viewer preferences? Only the
// presence of a non-empty catalog /ViewerPreferences dictionary is checked; its values are not read.
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
		prefs, err := pdfx.CatalogDict(ctx, "ViewerPreferences")
		return len(prefs) > 0, err
	})
}
