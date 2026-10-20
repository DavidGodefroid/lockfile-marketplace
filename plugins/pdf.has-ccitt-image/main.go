// Command pdf.has-ccitt-image answers: does the PDF hold a CCITT
// fax-encoded stream?
//
// Every object of the file is visited, orphans included. Only the /Filter
// entry of stream dicts is read; stream data is never decoded.
package main

import (
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

const ccitt = "CCITTFaxDecode"

// hasCCITT reports whether the stream dict d has /CCITTFaxDecode in its
// /Filter, a name or an array of names. Any other /Filter type is an error.
func hasCCITT(ctx *model.Context, d types.Dict) (bool, error) {
	o, ok := d.Find("Filter")
	if !ok {
		return false, nil
	}
	o, err := ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	switch v := o.(type) {
	case nil:
		return false, nil
	case types.Name:
		return v == ccitt, nil
	case types.Array:
		for _, x := range v {
			x, err := ctx.Dereference(x)
			if err != nil {
				return false, err
			}
			n, ok := x.(types.Name)
			if !ok {
				return false, fmt.Errorf("pdf.has-ccitt-image: /Filter array entry is %T, not a name", x)
			}
			if n == ccitt {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("pdf.has-ccitt-image: /Filter is %T, not a name or an array", o)
	}
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		var found bool
		var ferr error
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			s, ok := o.(types.StreamDict)
			if !ok {
				return false
			}
			found, ferr = hasCCITT(ctx, s.Dict)
			return found || ferr != nil
		})
		if err != nil {
			return false, err
		}
		return found, ferr
	})
}
