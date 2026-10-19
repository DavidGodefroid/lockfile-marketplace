package pdfx

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var (
	// ErrMissingMediaBox: no /MediaBox on the page nor any ancestor. pdfcpu
	// would default one; a page question must not answer on an invented box.
	ErrMissingMediaBox = errors.New("pdfx: page has no MediaBox")
	// ErrEmptyCropBox: the crop box does not intersect the media box.
	ErrEmptyCropBox = errors.New("pdfx: crop box outside the media box")
	// ErrBadBox: a page box that is not four numbers.
	ErrBadBox = errors.New("pdfx: malformed page box")
)

// maxTreeDepth bounds the walk up the page tree.
const maxTreeDepth = 64

// A4 in points, and the tolerance (1 mm) on each side.
const (
	A4Short     = 595.276
	A4Long      = 841.890
	A4Tolerance = 2.835
)

// Rect is a box in default user space units.
type Rect struct{ LLX, LLY, URX, URY float64 }

// Size returns the absolute width and height.
func (r Rect) Size() (w, h float64) {
	return math.Abs(r.URX - r.LLX), math.Abs(r.URY - r.LLY)
}

func (r Rect) norm() Rect {
	return Rect{min(r.LLX, r.URX), min(r.LLY, r.URY), max(r.LLX, r.URX), max(r.LLY, r.URY)}
}

// IsA4 reports whether w x h is A4 within 1 mm on each side, either
// orientation.
func IsA4(w, h float64) bool {
	short, long := min(w, h), max(w, h)
	return math.Abs(short-A4Short) <= A4Tolerance && math.Abs(long-A4Long) <= A4Tolerance
}

// inherited returns the first value of key on the page dict or its
// ancestors, read from the file (not pdfcpu's defaults).
func inherited(ctx *model.Context, page types.Dict, key string) (types.Object, error) {
	node := page
	for depth := 0; node != nil && depth < maxTreeDepth; depth++ {
		if o, ok := node.Find(key); ok && o != nil {
			return o, nil
		}
		p, ok := node.Find("Parent")
		if !ok {
			return nil, nil
		}
		next, err := ctx.DereferenceDict(p)
		if err != nil {
			return nil, err
		}
		node = next
	}
	return nil, nil
}

func rectOf(ctx *model.Context, o types.Object) (Rect, error) {
	a, err := ctx.DereferenceArray(o)
	if err != nil || len(a) != 4 {
		return Rect{}, ErrBadBox
	}
	var v [4]float64
	for i := range a {
		if v[i], err = ctx.DereferenceNumber(a[i]); err != nil {
			return Rect{}, ErrBadBox
		}
	}
	return Rect{v[0], v[1], v[2], v[3]}.norm(), nil
}

// PageBox returns the visible box of page (1-based): the inherited crop
// box clipped to the inherited media box, or the media box, scaled by the
// page's /UserUnit.
func PageBox(ctx *model.Context, page int) (box Rect, err error) {
	defer catch(&err)
	d, _, _, err := ctx.PageDict(context.Background(), page, false)
	if err != nil {
		return Rect{}, err
	}
	mo, err := inherited(ctx, d, "MediaBox")
	if err != nil {
		return Rect{}, err
	}
	if mo == nil {
		return Rect{}, ErrMissingMediaBox
	}
	media, err := rectOf(ctx, mo)
	if err != nil {
		return Rect{}, err
	}
	box = media
	co, err := inherited(ctx, d, "CropBox")
	if err != nil {
		return Rect{}, err
	}
	if co != nil {
		crop, err := rectOf(ctx, co)
		if err != nil {
			return Rect{}, err
		}
		box = Rect{max(media.LLX, crop.LLX), max(media.LLY, crop.LLY), min(media.URX, crop.URX), min(media.URY, crop.URY)}
		if box.URX <= box.LLX || box.URY <= box.LLY {
			return Rect{}, ErrEmptyCropBox
		}
	}
	unit := 1.0
	if o, ok := d.Find("UserUnit"); ok && o != nil {
		if unit, err = ctx.DereferenceNumber(o); err != nil || unit <= 0 {
			return Rect{}, fmt.Errorf("pdfx: invalid UserUnit")
		}
	}
	return Rect{box.LLX * unit, box.LLY * unit, box.URX * unit, box.URY * unit}, nil
}
