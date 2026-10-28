// Command pdf.has-hidden-annotation answers: does some page carry an
// annotation flagged hidden from display?
//
// Only annotations a page lists in /Annots count. Hidden is annotation
// flag bit 2 (Hidden) or bit 6 (NoView) of /F; an /F that is not an
// integer is an error, never 0.
package main

import (
	"fmt"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const (
	flagHidden = 1 << 1 // bit 2
	flagNoView = 1 << 5 // bit 6
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		var flagErr error
		err = pdfx.EachAnnotation(ctx, func(page int, a types.Dict) bool {
			fl, _, err := ctx.DereferenceIntegerEntry(a, "F")
			if err != nil {
				flagErr = fmt.Errorf("page %d: annotation /F: %w", page, err)
				return true
			}
			found = fl != nil && fl.Value()&(flagHidden|flagNoView) != 0
			return found
		})
		if err == nil {
			err = flagErr
		}
		if err != nil {
			return false, err
		}
		return found, nil
	})
}
