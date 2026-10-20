// Command pdf.has-jpeg2000-image answers: does the PDF hold a JPEG 2000
// (JPXDecode) stream?
//
// Every stream of the file is visited, orphans included. Stream data is
// never decoded: only the /Filter entry of the stream dict is read.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

var errFilter = errors.New("stream /Filter is neither a name nor an array of names")

// jpx reports whether the stream dict d has JPXDecode in its /Filter, a
// name or an array of names.
func jpx(ctx *model.Context, d types.Dict) (bool, error) {
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
		return v == "JPXDecode", nil
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
			if n == "JPXDecode" {
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
			s, ok := o.(types.StreamDict)
			if !ok {
				return false
			}
			found, ferr = jpx(ctx, s.Dict)
			return found || ferr != nil
		})
		if err != nil {
			return false, err
		}
		return found, ferr
	})
}
