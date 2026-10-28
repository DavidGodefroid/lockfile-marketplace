// Command pdf.has-associated-files answers: does the PDF declare associated files? Only the
// catalog /AF array (PDF/A-3) is read; the file specifications it points to are not.
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
		af, err := pdfx.CatalogArray(ctx, "AF")
		if err != nil {
			return false, err
		}
		return len(af) > 0, nil
	})
}
