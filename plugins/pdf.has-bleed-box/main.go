// Command pdf.has-bleed-box answers: does the PDF have at least one page
// with a /BleedBox?
package main

import (
	"context"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// errBadBleedBox: a /BleedBox that is not an array of four numbers.
var errBadBleedBox = errors.New("pdf.has-bleed-box: malformed /BleedBox")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		if ctx.PageCount == 0 {
			return false, pdfx.ErrNoPages
		}
		return hasBleedBox(ctx)
	})
}

// hasBleedBox reports whether a page dictionary has a /BleedBox. The key is
// not inheritable, so only the page dictionaries themselves are read.
func hasBleedBox(ctx *model.Context) (found bool, err error) {
	defer func() {
		if recover() != nil {
			found, err = false, pdfx.ErrParse
		}
	}()
	for p := 1; p <= ctx.PageCount; p++ {
		d, _, _, err := ctx.PageDict(context.Background(), p, false)
		if err != nil {
			return false, err
		}
		o, ok := d.Find("BleedBox")
		if !ok || o == nil {
			continue
		}
		a, err := ctx.DereferenceArray(o)
		if err != nil || len(a) != 4 {
			return false, errBadBleedBox
		}
		for _, n := range a {
			if _, err := ctx.DereferenceNumber(n); err != nil {
				return false, errBadBleedBox
			}
		}
		return true, nil
	}
	return false, nil
}
