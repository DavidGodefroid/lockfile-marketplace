// Command pdf.has-page-transition answers: does the PDF define a
// presentation transition or display duration for at least one page? Only
// the presence of a page's /Trans or /Dur is read; their values are not.
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
			for _, key := range []string{"Trans", "Dur"} {
				o, ok := d.Find(key)
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
		}
		return false, nil
	})
}
