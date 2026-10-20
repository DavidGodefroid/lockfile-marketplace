// Command pdf.has-auto-action answers: does the PDF declare an action that
// runs without a user gesture, on opening the document or a page?
//
// The catalog /OpenAction (other than a plain go-to), the catalog and page
// /AA, and the page-visibility triggers of annotations count. The actions
// are never run.
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
		return pdfx.HasAutoAction(ctx)
	})
}
