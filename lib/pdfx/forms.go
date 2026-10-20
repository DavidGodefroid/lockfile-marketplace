package pdfx

import (
	"errors"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ErrFieldTreeTooDeep is returned for a field tree deeper than
// maxTreeDepth.
var ErrFieldTreeTooDeep = errors.New("pdfx: field tree too deep")

// AcroForm returns the catalog's interactive form dictionary, or nil.
func AcroForm(ctx *model.Context) (form types.Dict, err error) {
	defer catch(&err)
	cat, err := ctx.Catalog()
	if err != nil {
		return nil, err
	}
	o, ok := cat.Find("AcroForm")
	if !ok {
		return nil, nil
	}
	return ctx.DereferenceDict(o)
}

// EachField calls fn on every node of the AcroForm field tree (from
// /Fields, down /Kids) with its field type, inherited from its ancestors
// ("" when no node sets /FT), until fn returns true. A node reached twice
// is visited once.
func EachField(ctx *model.Context, fn func(field types.Dict, ft string) (stop bool)) (err error) {
	defer catch(&err)
	form, err := AcroForm(ctx)
	if err != nil || form == nil {
		return err
	}
	o, ok := form.Find("Fields")
	if !ok {
		return nil
	}
	fields, err := ctx.DereferenceArray(o)
	if err != nil {
		return err
	}
	seen := map[int]bool{}
	_, err = eachField(ctx, fields, "", 0, seen, fn)
	return err
}

func eachField(ctx *model.Context, nodes types.Array, ft string, depth int, seen map[int]bool, fn func(types.Dict, string) bool) (bool, error) {
	if depth > maxTreeDepth {
		return false, ErrFieldTreeTooDeep
	}
	for _, n := range nodes {
		if ref, ok := n.(types.IndirectRef); ok {
			num := ref.ObjectNumber.Value()
			if seen[num] {
				continue
			}
			seen[num] = true
		}
		d, err := ctx.DereferenceDict(n)
		if err != nil {
			return false, err
		}
		if d == nil {
			continue
		}
		t := ft
		if name := d.NameEntry("FT"); name != nil {
			t = *name
		}
		if fn(d, t) {
			return true, nil
		}
		k, ok := d.Find("Kids")
		if !ok {
			continue
		}
		kids, err := ctx.DereferenceArray(k)
		if err != nil {
			return false, err
		}
		if stop, err := eachField(ctx, kids, t, depth+1, seen, fn); err != nil || stop {
			return stop, err
		}
	}
	return false, nil
}

// SignatureValues returns the signature dictionaries of the field tree:
// the /V of each signature field, document time stamps included, when it
// has a /ByteRange and a /Contents. Nothing is verified.
func SignatureValues(ctx *model.Context) (sigs []types.Dict, err error) {
	defer catch(&err)
	var derr error
	err = EachField(ctx, func(field types.Dict, ft string) bool {
		if ft != "Sig" {
			return false
		}
		o, ok := field.Find("V")
		if !ok || o == nil {
			return false
		}
		v, e := ctx.DereferenceDict(o)
		if e != nil {
			derr = e
			return true
		}
		if v == nil {
			return false
		}
		_, br := v.Find("ByteRange")
		_, c := v.Find("Contents")
		if br && c {
			sigs = append(sigs, v)
		}
		return false
	})
	if err == nil {
		err = derr
	}
	return sigs, err
}
