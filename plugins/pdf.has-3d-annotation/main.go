// Command pdf.has-3d-annotation answers: does some page carry a 3D
// annotation (an embedded 3D model)?
//
// Only annotations a page lists in /Annots count; the subtype must be 3D
// (ISO 32000-2 §13.6.2).
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
		found := false
		err = pdfx.EachAnnotation(ctx, func(_ int, a types.Dict) bool {
			s := a.NameEntry("Subtype")
			found = s != nil && *s == "3D"
			return found
		})
		return found, err
	})
}
