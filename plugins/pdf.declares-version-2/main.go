// Command pdf.declares-version-2 answers: does the PDF declare PDF version 2.0? The effective version
// is the higher of the header and the catalog /Version.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return ctx.XRefTable.Version() >= model.V20, nil
	})
}
