// Command pdf.has-read-only-field answers: does the PDF have a read-only
// form field?
//
// A read-only field is a node of the AcroForm field tree whose field
// flags (/Ff) have bit 1 (ReadOnly) set. A kid that inherits /Ff gets it
// from a node that is itself visited, so inheritance needs no extra walk.
// Field values are never read; a /Ff that is not an integer is an error.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// readOnly is field flag bit 1 (ISO 32000-1, table 221).
const readOnly = 1 << 0

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return hasReadOnlyField(ctx)
	})
}

func hasReadOnlyField(ctx *model.Context) (bool, error) {
	found := false
	var ferr error
	err := pdfx.EachField(ctx, func(field types.Dict, _ string) bool {
		o, ok := field.Find("Ff")
		if !ok || o == nil {
			return false
		}
		ff, e := ctx.DereferenceInteger(o)
		if e != nil {
			ferr = e
			return true
		}
		found = ff != nil && ff.Value()&readOnly != 0
		return found
	})
	if err == nil {
		err = ferr
	}
	return found && err == nil, err
}
