// Command pdf.has-javascript answers: does the PDF contain JavaScript?
//
// Every object of the file is visited, not only the reachable actions, so
// orphan actions and actions chained by /Next count. The script itself is
// never read. XFA form scripts are out of scope.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// isJavaScript reports whether o is a JavaScript action or has a /JS entry.
func isJavaScript(o types.Object) bool {
	d, ok := o.(types.Dict)
	if !ok {
		return false
	}
	if _, ok := d.Find("JS"); ok {
		return true
	}
	s := d.NameEntry("S")
	return s != nil && *s == "JavaScript"
}

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
		if o, ok := cat.Find("Names"); ok {
			names, err := ctx.DereferenceDict(o)
			if err != nil {
				return false, err
			}
			if _, ok := names.Find("JavaScript"); ok {
				return true, nil
			}
		}
		found := false
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			found = isJavaScript(o)
			return found
		})
		return found, err
	})
}
