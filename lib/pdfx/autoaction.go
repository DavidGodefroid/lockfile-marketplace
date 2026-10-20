package pdfx

import (
	"context"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// annotTriggers are the annotation /AA keys a viewer fires without a user
// gesture: page opened, closed, made visible, made invisible.
var annotTriggers = []string{"PO", "PC", "PV", "PI"}

// HasAutoAction reports whether the file declares an action a viewer runs
// without a user gesture: a catalog /OpenAction that is an action dict
// other than a plain /S /GoTo without /Next (a destination array does not
// count), a non-empty catalog or page /AA, or a page annotation's /AA with
// a PO, PC, PV or PI key. The actions themselves are never run nor
// resolved further.
func HasAutoAction(ctx *model.Context) (found bool, err error) {
	defer catch(&err)
	cat, err := ctx.Catalog()
	if err != nil {
		return false, err
	}
	if found, err = autoOpenAction(ctx, cat); err != nil || found {
		return found, err
	}
	if found, err = hasTrigger(ctx, cat, nil); err != nil || found {
		return found, err
	}
	for p := 1; p <= ctx.PageCount; p++ {
		d, _, _, err := ctx.PageDict(context.Background(), p, false)
		if err != nil {
			return false, err
		}
		if d == nil {
			continue
		}
		if found, err = hasTrigger(ctx, d, nil); err != nil || found {
			return found, err
		}
		if found, err = annotsTrigger(ctx, d); err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func autoOpenAction(ctx *model.Context, cat types.Dict) (bool, error) {
	o, ok := cat.Find("OpenAction")
	if !ok || o == nil {
		return false, nil
	}
	o, err := ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	d, ok := o.(types.Dict)
	if !ok {
		return false, nil
	}
	s := d.NameEntry("S")
	_, next := d.Find("Next")
	return s == nil || *s != "GoTo" || next, nil
}

// hasTrigger reports whether d has an /AA dict holding one of keys, or any
// entry when keys is nil.
func hasTrigger(ctx *model.Context, d types.Dict, keys []string) (bool, error) {
	o, ok := d.Find("AA")
	if !ok || o == nil {
		return false, nil
	}
	aa, err := ctx.DereferenceDict(o)
	if err != nil || aa == nil {
		return false, err
	}
	if keys == nil {
		return len(aa) > 0, nil
	}
	for _, k := range keys {
		if v, ok := aa.Find(k); ok && v != nil {
			return true, nil
		}
	}
	return false, nil
}

func annotsTrigger(ctx *model.Context, page types.Dict) (bool, error) {
	o, ok := page.Find("Annots")
	if !ok || o == nil {
		return false, nil
	}
	annots, err := ctx.DereferenceArray(o)
	if err != nil {
		return false, err
	}
	for _, a := range annots {
		d, err := ctx.DereferenceDict(a)
		if err != nil {
			return false, err
		}
		if d == nil {
			continue
		}
		if found, err := hasTrigger(ctx, d, annotTriggers); err != nil || found {
			return found, err
		}
	}
	return false, nil
}
