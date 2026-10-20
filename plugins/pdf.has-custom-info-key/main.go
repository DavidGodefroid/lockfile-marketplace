// Command pdf.has-custom-info-key answers: does the PDF document information dictionary have a
// key outside the standard ones? Neither the key names nor their values are returned or kept:
// only whether a non-standard key exists.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// standard lists the document information keys defined by ISO 32000 (14.3.3).
var standard = map[string]bool{
	"Title":        true,
	"Author":       true,
	"Subject":      true,
	"Keywords":     true,
	"Creator":      true,
	"Producer":     true,
	"CreationDate": true,
	"ModDate":      true,
	"Trapped":      true,
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return hasCustomKey(ctx)
	})
}

// hasCustomKey reports whether the trailer /Info dictionary has a non-null
// entry whose key is not a standard one. No /Info, or a null one, is false.
// An /Info that is not a dictionary is pdfx.ErrWrongType; pdfx.Open (pdfcpu,
// relaxed) already drops such an /Info, so that branch is defensive.
func hasCustomKey(ctx *model.Context) (ok bool, err error) {
	defer func() {
		if p := recover(); p != nil {
			ok, err = false, pdfx.ErrParse
		}
	}()
	if ctx.Info == nil {
		return false, nil
	}
	info, err := ctx.DereferenceDict(*ctx.Info)
	if err != nil {
		return false, pdfx.ErrWrongType
	}
	for k, v := range info {
		if standard[k] || v == nil {
			continue
		}
		o, err := ctx.Dereference(v)
		if err != nil {
			return false, err
		}
		if o != nil {
			return true, nil
		}
	}
	return false, nil
}
