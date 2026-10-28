// Command pdf.has-transparency answers: does the PDF use transparency?
//
// Every object of the file is visited. A graphics state counts when it has
// a soft mask other than /None, a stroking or non-stroking alpha (/CA, /ca)
// below 1, or a blend mode other than Normal or Compatible; graphics states
// are found as /Type /ExtGState dicts and as the entries of any resources
// /ExtGState dict (the type key is optional). An image XObject counts when
// it has an /SMask. Whether the state or image is actually painted is not
// checked. A malformed entry is an error, never 0.
package main

import (
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// isName reports whether d's entry key is the name v.
func isName(d types.Dict, key, v string) bool {
	n := d.NameEntry(key)
	return n != nil && *n == v
}

// transparentState reports whether the graphics state gs uses transparency.
func transparentState(ctx *model.Context, gs types.Dict) (bool, error) {
	if o, ok := gs.Find("SMask"); ok {
		o, err := ctx.Dereference(o)
		if err != nil {
			return false, err
		}
		if n, ok := o.(types.Name); !ok || n != "None" {
			return true, nil
		}
	}
	for _, key := range []string{"CA", "ca"} {
		if o, ok := gs.Find(key); ok {
			a, err := ctx.DereferenceNumber(o)
			if err != nil {
				return false, fmt.Errorf("ExtGState /%s: %w", key, err)
			}
			if a < 1 {
				return true, nil
			}
		}
	}
	if o, ok := gs.Find("BM"); ok {
		return blends(ctx, o)
	}
	return false, nil
}

// blends reports whether the blend mode o (a name, or an array of names
// in preference order) names a mode other than Normal or Compatible.
func blends(ctx *model.Context, o types.Object) (bool, error) {
	o, err := ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	modes, ok := o.(types.Array)
	if !ok {
		modes = types.Array{o}
	}
	for _, m := range modes {
		m, err := ctx.Dereference(m)
		if err != nil {
			return false, err
		}
		n, ok := m.(types.Name)
		if !ok {
			return false, fmt.Errorf("ExtGState /BM: not a name: %v", m)
		}
		if n != "Normal" && n != "Compatible" {
			return true, nil
		}
	}
	return false, nil
}

// transparent reports whether the dict d, met while walking every object,
// is or holds a transparent graphics state, or is a soft-masked image.
func transparent(ctx *model.Context, d types.Dict) (bool, error) {
	if isName(d, "Subtype", "Image") {
		_, ok := d.Find("SMask")
		return ok, nil
	}
	if isName(d, "Type", "ExtGState") {
		if found, err := transparentState(ctx, d); err != nil || found {
			return found, err
		}
	}
	o, ok := d.Find("ExtGState")
	if !ok {
		return false, nil
	}
	states, err := ctx.DereferenceDict(o)
	if err != nil {
		return false, err
	}
	for _, s := range states {
		gs, err := ctx.DereferenceDict(s)
		if err != nil {
			return false, err
		}
		if found, err := transparentState(ctx, gs); err != nil || found {
			return found, err
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
		var werr error
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			d, ok := o.(types.Dict)
			if !ok {
				return false
			}
			found, werr = transparent(ctx, d)
			return found || werr != nil
		})
		if err != nil {
			return false, err
		}
		return found, werr
	})
}
