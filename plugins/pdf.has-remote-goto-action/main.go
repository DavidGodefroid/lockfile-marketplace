// Command pdf.has-remote-goto-action answers: does the PDF contain a
// remote go-to action, one that jumps into another PDF file (GoToR) or
// into an embedded file (GoToE)?
//
// Every object of the file is visited, so orphan actions and actions
// chained by /Next count. The target file and destination are never read.
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
		return pdfx.HasAction(ctx, "GoToR", "GoToE")
	})
}
