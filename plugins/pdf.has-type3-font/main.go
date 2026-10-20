// Command pdf.has-type3-font answers: does the PDF contain a Type 3 font?
//
// Every object of the file is visited, orphans included. Glyph procedures
// and font names are never read.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func isName(d types.Dict, key, value string) bool {
	n := d.NameEntry(key)
	return n != nil && *n == value
}

// isType3 reports whether o is a font dictionary of subtype Type3.
func isType3(o types.Object) bool {
	d, ok := o.(types.Dict)
	return ok && isName(d, "Type", "Font") && isName(d, "Subtype", "Type3")
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			found = isType3(o)
			return found
		})
		return found, err
	})
}
