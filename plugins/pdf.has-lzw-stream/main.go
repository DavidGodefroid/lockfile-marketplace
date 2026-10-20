// Command pdf.has-lzw-stream answers: does the PDF have an LZW-compressed
// stream?
//
// Every object of the file is visited, orphans included. Only the /Filter
// entry of stream dictionaries is read; stream data is never decoded.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

const lzw = "LZWDecode"

// errFilter guards a malformed /Filter. pdfcpu already drops such streams
// while reading; this keeps one from ever answering 0.
var errFilter = errors.New("pdf.has-lzw-stream: /Filter is neither a name nor an array of names")

// usesLZW reports whether the stream dict d lists LZWDecode in its /Filter.
func usesLZW(ctx *model.Context, d types.Dict) (bool, error) {
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
		return v == lzw, nil
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
			if n == lzw {
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
			found, ferr = usesLZW(ctx, sd.Dict)
			return found || ferr != nil
		})
		if err == nil {
			err = ferr
		}
		return found, err
	})
}
