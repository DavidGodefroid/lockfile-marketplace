package pdfx

import (
	"context"
	"errors"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Drawing is what a page's content draws, form XObjects included.
type Drawing struct {
	// Text: a text-showing operator (Tj TJ ' "), in any rendering mode,
	// invisible text included.
	Text bool
	// Image: an image XObject drawn with Do, or an inline image (BI).
	Image bool
}

// PageDraws reports what page (1-based) draws: its content streams, then
// each form XObject drawn with Do, recursively (up to maxFormDepth deep,
// each object once). A form without /Resources uses the resources of the
// stream that draws it. A Do whose name is not in the resources draws
// nothing. Annotations and Type3 glyph procedures are not considered.
func PageDraws(ctx *model.Context, page int) (dr Drawing, err error) {
	defer catch(&err)
	d, _, inh, err := ctx.PageDict(context.Background(), page, false)
	if err != nil {
		return Drawing{}, err
	}
	content, err := ctx.PageContent(d, page)
	if errors.Is(err, model.ErrNoContent) {
		return Drawing{}, nil
	}
	if err != nil {
		return Drawing{}, err
	}
	var res types.Dict
	if inh != nil {
		res = inh.Resources
	}
	if o, ok := d.Find("Resources"); ok {
		if r, err := ctx.DereferenceDict(o); err == nil && r != nil {
			res = r
		}
	}
	w := drawWalker{ctx: ctx, seen: map[int]bool{}}
	err = w.scan(content, res, 0)
	return w.dr, err
}

type drawWalker struct {
	ctx  *model.Context
	seen map[int]bool
	dr   Drawing
}

func (w *drawWalker) done() bool { return w.dr.Text && w.dr.Image }

func (w *drawWalker) scan(content []byte, res types.Dict, depth int) error {
	var err error
	lex(content, func(op string, args []token) bool {
		switch op {
		case "Tj", "TJ", "'", "\"":
			w.dr.Text = true
		case "BI":
			w.dr.Image = true
		case "Do":
			if len(args) == 1 && args[0].kind == tokName {
				err = w.do(args[0].name, res, depth)
			}
		}
		return err == nil && !w.done()
	})
	return err
}

// do follows the XObject called name in res.
func (w *drawWalker) do(name string, res types.Dict, depth int) error {
	if res == nil {
		return nil
	}
	xo, ok := res.Find("XObject")
	if !ok {
		return nil
	}
	xobjs, err := w.ctx.DereferenceDict(xo)
	if err != nil || xobjs == nil {
		return err
	}
	o, ok := xobjs.Find(name)
	if !ok || o == nil {
		return nil
	}
	n := -1
	if ref, ok := o.(types.IndirectRef); ok {
		n = ref.ObjectNumber.Value()
		if w.seen[n] {
			return nil
		}
	}
	sd, _, err := w.ctx.DereferenceStreamDict(o)
	if err != nil || sd == nil {
		return err
	}
	sub := sd.Dict.NameEntry("Subtype")
	if sub == nil {
		return nil
	}
	switch *sub {
	case "Image":
		w.dr.Image = true
		return nil
	case "Form":
	default:
		return nil
	}
	if depth >= maxFormDepth {
		return nil
	}
	// Marked once scanned, so a form first met too deep is still
	// followed from a shallower Do.
	if n >= 0 {
		w.seen[n] = true
	}
	if err := sd.Decode(); err != nil {
		return err
	}
	subRes := res
	if r, ok := sd.Dict.Find("Resources"); ok {
		d, err := w.ctx.DereferenceDict(r)
		if err != nil {
			return err
		}
		if d != nil {
			subRes = d
		}
	}
	return w.scan(sd.Content, subRes, depth+1)
}
