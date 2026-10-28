// Command pdf.has-figure-without-alt answers: does the PDF's structure tree
// have a Figure element without alternate text? Only the presence of /Alt
// and /ActualText is checked, never their text.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachStructElem(ctx, func(elem types.Dict, role string) bool {
			found = role == "Figure" && !has(elem, "Alt") && !has(elem, "ActualText")
			return found
		})
		return found, err
	})
}

// has reports whether d has key with a non-null value.
func has(d types.Dict, key string) bool {
	o, ok := d.Find(key)
	return ok && o != nil
}
