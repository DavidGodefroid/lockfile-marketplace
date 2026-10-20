package pdfx

import (
	"context"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// EachAnnotation calls fn on every annotation dictionary listed in a
// page's /Annots, page by page, until fn returns true. Annotations no page
// lists are not visited; a null entry is skipped.
func EachAnnotation(ctx *model.Context, fn func(page int, annot types.Dict) (stop bool)) (err error) {
	defer catch(&err)
	for p := 1; p <= ctx.PageCount; p++ {
		d, _, _, err := ctx.PageDict(context.Background(), p, false)
		if err != nil {
			return err
		}
		o, ok := d.Find("Annots")
		if !ok {
			continue
		}
		annots, err := ctx.DereferenceArray(o)
		if err != nil {
			return err
		}
		for _, a := range annots {
			annot, err := ctx.DereferenceDict(a)
			if err != nil {
				return err
			}
			if annot != nil && fn(p, annot) {
				return nil
			}
		}
	}
	return nil
}
