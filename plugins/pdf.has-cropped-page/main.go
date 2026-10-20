// Command pdf.has-cropped-page answers: does the PDF have at least one page
// whose CropBox differs from its MediaBox by more than 1 mm on any edge?
package main

import (
	"context"
	"math"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// maxTreeDepth bounds the walk up the page tree.
const maxTreeDepth = 64

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
			cropped, err := pageCropped(ctx, p)
			if err != nil {
				return false, err
			}
			if cropped {
				return true, nil
			}
		}
		return false, nil
	})
}

// pageCropped reports whether page p has an effective CropBox whose
// intersection with the MediaBox moves an edge of the MediaBox by more
// than 1 mm. Both boxes are compared after /UserUnit scaling.
func pageCropped(ctx *model.Context, p int) (bool, error) {
	// PageBox validates both boxes and returns the crop clipped to the media.
	visible, err := pdfx.PageBox(ctx, p)
	if err != nil {
		return false, err
	}
	d, _, _, err := ctx.PageDict(context.Background(), p, false)
	if err != nil {
		return false, err
	}
	crop, err := inherited(ctx, d, "CropBox")
	if err != nil || crop == nil {
		return false, err
	}
	mo, err := inherited(ctx, d, "MediaBox")
	if err != nil {
		return false, err
	}
	media, err := rectOf(ctx, mo)
	if err != nil {
		return false, err
	}
	unit := 1.0
	if o, ok := d.Find("UserUnit"); ok && o != nil {
		if unit, err = ctx.DereferenceNumber(o); err != nil {
			return false, err
		}
	}
	edges := [4][2]float64{
		{visible.LLX, media.LLX * unit},
		{visible.LLY, media.LLY * unit},
		{visible.URX, media.URX * unit},
		{visible.URY, media.URY * unit},
	}
	for _, e := range edges {
		if math.Abs(e[0]-e[1]) > pdfx.A4Tolerance {
			return true, nil
		}
	}
	return false, nil
}

// inherited returns the first value of key on the page dict or its
// ancestors, as written in the file.
func inherited(ctx *model.Context, page types.Dict, key string) (types.Object, error) {
	node := page
	for depth := 0; node != nil && depth < maxTreeDepth; depth++ {
		if o, ok := node.Find(key); ok && o != nil {
			return o, nil
		}
		parent, ok := node.Find("Parent")
		if !ok {
			return nil, nil
		}
		next, err := ctx.DereferenceDict(parent)
		if err != nil {
			return nil, err
		}
		node = next
	}
	return nil, nil
}

// rectOf reads a four-number box, normalised so LL is the lower-left.
func rectOf(ctx *model.Context, o types.Object) (pdfx.Rect, error) {
	a, err := ctx.DereferenceArray(o)
	if err != nil || len(a) != 4 {
		return pdfx.Rect{}, pdfx.ErrBadBox
	}
	var v [4]float64
	for i := range a {
		if v[i], err = ctx.DereferenceNumber(a[i]); err != nil {
			return pdfx.Rect{}, pdfx.ErrBadBox
		}
	}
	return pdfx.Rect{
		LLX: min(v[0], v[2]), LLY: min(v[1], v[3]),
		URX: max(v[0], v[2]), URY: max(v[1], v[3]),
	}, nil
}
