// Command pdf.has-jbig2-image answers: does the PDF hold a JBIG2-encoded
// stream?
//
// Every object of the file is visited, orphans included. Only the /Filter
// entry of stream dicts is read; stream data is never decoded.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

const jbig2 = "JBIG2Decode"

var errFilter = errors.New("pdf.has-jbig2-image: /Filter is neither a name nor an array of names")

// usesJBIG2 reports whether the filter chain f, a name or an array of
// names, includes JBIG2Decode.
func usesJBIG2(ctx *model.Context, f types.Object) (bool, error) {
	f, err := ctx.Dereference(f)
	if err != nil {
		return false, err
	}
	switch v := f.(type) {
	case types.Name:
		return v.Value() == jbig2, nil
	case types.Array:
		for _, x := range v {
			x, err := ctx.Dereference(x)
			if err != nil {
				return false, err
			}
			n, ok := x.(types.Name)
			if !ok {
				return false, errFilter
			}
			if n.Value() == jbig2 {
				return true, nil
			}
		}
		return false, nil
	}
	return false, errFilter
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		var ferr error
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			sd, ok := o.(types.StreamDict)
			if !ok {
				return false
			}
			filter, ok := sd.Find("Filter")
			if !ok || filter == nil {
				return false
			}
			found, ferr = usesJBIG2(ctx, filter)
			return found || ferr != nil
		})
		if err == nil {
			err = ferr
		}
		return found, err
	})
}
