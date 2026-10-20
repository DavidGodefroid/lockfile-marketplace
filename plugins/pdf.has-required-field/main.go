// Command pdf.has-required-field answers: does the PDF have a required form
// field?
//
// A field is required when bit 2 (Required) is set in the field flags (/Ff)
// of a node of the AcroForm field tree. A /Ff that is not an integer is an
// error; field values are never read.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// required is bit 2 of the field flags (ISO 32000-1, table 221).
const required = 1 << 1

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		var ffErr error
		err = pdfx.EachField(ctx, func(field types.Dict, _ string) bool {
			ff, _, err := ctx.DereferenceIntegerEntry(field, "Ff")
			if err != nil {
				ffErr = err
				return true
			}
			found = ff != nil && ff.Value()&required != 0
			return found
		})
		if err == nil {
			err = ffErr
		}
		if err != nil {
			return false, err
		}
		return found, nil
	})
}
