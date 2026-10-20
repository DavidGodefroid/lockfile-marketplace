// Command pdf.has-calculated-field answers: does the PDF form declare a
// field calculation order?
//
// The calculation order is the /CO array of the catalog's interactive
// form dictionary: the fields whose value is computed from others, in the
// order they are recalculated. The calculation scripts are never read.
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
		form, err := pdfx.AcroForm(ctx)
		if err != nil || form == nil {
			return false, err
		}
		o, ok := form.Find("CO")
		if !ok {
			return false, nil
		}
		co, err := ctx.DereferenceArray(o)
		if err != nil {
			return false, err
		}
		return len(co) > 0, nil
	})
}
