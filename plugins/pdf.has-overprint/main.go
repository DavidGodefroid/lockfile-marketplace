// Command pdf.has-overprint answers: does the PDF turn on overprint?
//
// Every object of the file is visited, orphans and inline dicts included.
// /OP and /op are graphics state parameter keys, defined only for ExtGState
// dicts, so any dict carrying one of them is read as an ExtGState. Content
// streams are never decoded.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

var errOverprint = errors.New("pdf.has-overprint: ExtGState /OP or /op is not a boolean")

// overprint reports whether the dict d has /OP true or /op true.
func overprint(ctx *model.Context, d types.Dict) (bool, error) {
	for _, key := range []string{"OP", "op"} {
		o, ok := d.Find(key)
		if !ok {
			continue
		}
		o, err := ctx.Dereference(o)
		if err != nil {
			return false, err
		}
		b, ok := o.(types.Boolean)
		if !ok {
			return false, errOverprint
		}
		if b {
			return true, nil
		}
	}
	return false, nil
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
			d, ok := o.(types.Dict)
			if !ok {
				return false
			}
			found, ferr = overprint(ctx, d)
			return found || ferr != nil
		})
		if err == nil {
			err = ferr
		}
		return found, err
	})
}
