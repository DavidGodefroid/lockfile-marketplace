// Command pdf.has-thumbnails answers: does the PDF embed a thumbnail image
// for at least one page? Only the presence of a page's /Thumb is read; the
// image itself is not decoded.
package main

import (
	"context"
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
		if ctx.PageCount == 0 {
			return false, pdfx.ErrNoPages
		}
		for p := 1; p <= ctx.PageCount; p++ {
			d, _, _, err := ctx.PageDict(context.Background(), p, false)
			if err != nil {
				return false, err
			}
			o, ok := d.Find("Thumb")
			if !ok {
				continue
			}
			o, err = ctx.Dereference(o)
			if err != nil {
				return false, err
			}
			if o != nil {
				return true, nil
			}
		}
		return false, nil
	})
}
