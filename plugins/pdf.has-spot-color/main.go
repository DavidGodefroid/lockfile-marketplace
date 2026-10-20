// Command pdf.has-spot-color answers: does the PDF use a spot colour space?
//
// Every object of the file is visited, so a Separation or DeviceN colour
// space counts wherever it is declared (page resources, forms, images,
// shadings, orphan objects). Colorant names are never read.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// isSpotColorSpace reports whether o is a [/Separation …] or [/DeviceN …]
// colour space array.
func isSpotColorSpace(o types.Object) bool {
	a, ok := o.(types.Array)
	if !ok || len(a) == 0 {
		return false
	}
	n, ok := a[0].(types.Name)
	return ok && (n == "Separation" || n == "DeviceN")
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			found = isSpotColorSpace(o)
			return found
		})
		return found, err
	})
}
