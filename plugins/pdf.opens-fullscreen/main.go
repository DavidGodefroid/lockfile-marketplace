// Command pdf.opens-fullscreen answers: does the PDF ask the viewer to open
// in full-screen mode? Only the catalog /PageMode entry is read.
package main

import (
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var errNotName = errors.New("pdf.opens-fullscreen: /PageMode is not a name")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		cat, err := ctx.Catalog()
		if err != nil {
			return false, err
		}
		o, found := cat.Find("PageMode")
		if !found || o == nil {
			return false, nil
		}
		o, err = ctx.Dereference(o)
		if err != nil || o == nil {
			return false, err
		}
		mode, ok := o.(types.Name)
		if !ok {
			return false, errNotName
		}
		return mode.Value() == "FullScreen", nil
	})
}
