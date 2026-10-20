// Command pdf.has-icc-color-space answers: does the PDF use an ICC-based
// colour space?
//
// Every object of the file is visited, and every array nested in it, so a
// colour space inside /Indexed, /DeviceN or a pattern counts as well as one
// in page resources. The ICC profile stream is never read. An output
// intent's /DestOutputProfile alone is not a colour space and does not count.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// isICCBased reports whether o is an array [/ICCBased …].
func isICCBased(o types.Object) bool {
	a, ok := o.(types.Array)
	if !ok || len(a) == 0 {
		return false
	}
	n, ok := a[0].(types.Name)
	return ok && n == "ICCBased"
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			found = isICCBased(o)
			return found
		})
		return found, err
	})
}
