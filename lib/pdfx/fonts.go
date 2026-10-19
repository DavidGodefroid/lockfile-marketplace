package pdfx

import (
	"context"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// maxFormDepth bounds the descent into form XObjects.
const maxFormDepth = 16

// UsesUnembeddedFont reports whether a font referenced from a page's
// resources (or from a form XObject's, recursively) has no FontFile,
// FontFile2 or FontFile3. A font without a descriptor (the standard 14)
// counts as not embedded; a Type3 font draws its own glyphs and counts as
// embedded; a Type0 font is judged by its descendant.
func UsesUnembeddedFont(ctx *model.Context) (found bool, err error) {
	defer catch(&err)
	if ctx.PageCount == 0 {
		return false, ErrNoPages
	}
	seen := map[int]bool{}
	for p := 1; p <= ctx.PageCount; p++ {
		d, _, inh, err := ctx.PageDict(context.Background(), p, false)
		if err != nil {
			return false, err
		}
		res := inh.Resources
		if o, ok := d.Find("Resources"); ok {
			if r, err := ctx.DereferenceDict(o); err == nil && r != nil {
				res = r
			}
		}
		found, err := resourcesUnembedded(ctx, res, 0, seen)
		if err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func resourcesUnembedded(ctx *model.Context, res types.Dict, depth int, seen map[int]bool) (bool, error) {
	if res == nil || depth > maxFormDepth {
		return false, nil
	}
	if o, ok := res.Find("Font"); ok {
		fonts, err := ctx.DereferenceDict(o)
		if err != nil {
			return false, err
		}
		for _, f := range fonts {
			font, err := ctx.DereferenceDict(f)
			if err != nil {
				return false, err
			}
			if font == nil {
				continue
			}
			un, err := fontUnembedded(ctx, font)
			if err != nil || un {
				return un, err
			}
		}
	}
	if o, ok := res.Find("XObject"); ok {
		xobjs, err := ctx.DereferenceDict(o)
		if err != nil {
			return false, err
		}
		for _, x := range xobjs {
			if ref, ok := x.(types.IndirectRef); ok {
				n := ref.ObjectNumber.Value()
				if seen[n] {
					continue
				}
				seen[n] = true
			}
			obj, err := ctx.Dereference(x)
			if err != nil {
				return false, err
			}
			sd, ok := obj.(types.StreamDict)
			if !ok || sd.Dict.NameEntry("Subtype") == nil || *sd.Dict.NameEntry("Subtype") != "Form" {
				continue
			}
			sub, ok := sd.Dict.Find("Resources")
			if !ok {
				continue
			}
			subRes, err := ctx.DereferenceDict(sub)
			if err != nil {
				return false, err
			}
			found, err := resourcesUnembedded(ctx, subRes, depth+1, seen)
			if err != nil || found {
				return found, err
			}
		}
	}
	return false, nil
}

func fontUnembedded(ctx *model.Context, font types.Dict) (bool, error) {
	sub := font.NameEntry("Subtype")
	if sub != nil && *sub == "Type3" {
		return false, nil
	}
	if sub != nil && *sub == "Type0" {
		o, ok := font.Find("DescendantFonts")
		if !ok {
			return true, nil
		}
		arr, err := ctx.DereferenceArray(o)
		if err != nil {
			return false, err
		}
		if len(arr) == 0 {
			return true, nil
		}
		desc, err := ctx.DereferenceDict(arr[0])
		if err != nil {
			return false, err
		}
		if desc == nil {
			return true, nil
		}
		font = desc
	}
	o, ok := font.Find("FontDescriptor")
	if !ok {
		return true, nil
	}
	fd, err := ctx.DereferenceDict(o)
	if err != nil {
		return false, err
	}
	if fd == nil {
		return true, nil
	}
	for _, k := range []string{"FontFile", "FontFile2", "FontFile3"} {
		if _, ok := fd.Find(k); ok {
			return false, nil
		}
	}
	return true, nil
}
