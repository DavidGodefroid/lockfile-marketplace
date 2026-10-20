// Command pdf.has-external-link answers: does the PDF link outside itself,
// to a URI or to another file?
//
// Every object of the file is visited, so orphan actions and actions
// chained by /Next count. The target is never read. Launch actions are
// pdf.has-launch-action's question.
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
		return pdfx.HasAction(ctx, "URI", "GoToR")
	})
}
