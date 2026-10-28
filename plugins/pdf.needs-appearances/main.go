// Command pdf.needs-appearances answers: does the PDF ask the viewer to
// regenerate its form field appearances?
//
// The flag is /NeedAppearances true in the catalog's interactive form
// dictionary. A /NeedAppearances that is not a boolean is malformed and
// is an error; field values are never read.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
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
		o, ok := form.Find("NeedAppearances")
		if !ok {
			return false, nil
		}
		o, err = ctx.Dereference(o)
		if err != nil {
			return false, err
		}
		if o == nil {
			return false, nil
		}
		b, ok := o.(types.Boolean)
		if !ok {
			return false, pdfx.ErrParse
		}
		return b.Value(), nil
	})
}
