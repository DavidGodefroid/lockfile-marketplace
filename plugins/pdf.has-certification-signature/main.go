// Command pdf.has-certification-signature answers: does the PDF carry a
// certification signature? Only the catalog /Perms /DocMDP entry is read; the
// signature is neither verified nor matched against the form fields.
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
		perms, err := pdfx.CatalogDict(ctx, "Perms")
		if err != nil || perms == nil {
			return false, err
		}
		o, ok := perms.Find("DocMDP")
		if !ok || o == nil {
			return false, nil
		}
		// A /DocMDP that is not a dictionary is malformed: an error, never 0.
		sig, err := ctx.DereferenceDict(o)
		if err != nil {
			return false, err
		}
		return sig != nil, nil
	})
}
