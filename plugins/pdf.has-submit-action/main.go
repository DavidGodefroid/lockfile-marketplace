// Command pdf.has-submit-action answers: does the PDF contain a
// submit-form action, one that sends form data to a destination?
//
// Every object of the file is visited, so orphan actions and actions
// chained by /Next count. The destination is never read.
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
		return pdfx.HasAction(ctx, "SubmitForm")
	})
}
