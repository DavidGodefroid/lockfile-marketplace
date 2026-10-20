// Command pdf.annotations-lack-tab-order answers: does some page with
// annotations lack the structure tab order PDF/UA requires?
//
// PDF/UA-1 (ISO 14289-1 §7.18.3) requires /Tabs /S on every page that
// lists at least one annotation in /Annots. /Tabs is not inheritable, so
// only the page dictionary itself is read.
package main

import (
	"context"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var errBadTabs = errors.New("pdf.annotations-lack-tab-order: /Tabs is not a name")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		if ctx.PageCount == 0 {
			return false, pdfx.ErrNoPages
		}
		var pages []int
		err = pdfx.EachAnnotation(ctx, func(p int, _ types.Dict) bool {
			if len(pages) == 0 || pages[len(pages)-1] != p {
				pages = append(pages, p)
			}
			return false
		})
		if err != nil {
			return false, err
		}
		for _, p := range pages {
			ok, err := structureOrder(ctx, p)
			if err != nil {
				return false, err
			}
			if !ok {
				return true, nil
			}
		}
		return false, nil
	})
}

// structureOrder reports whether page p has /Tabs /S.
func structureOrder(ctx *model.Context, p int) (bool, error) {
	d, _, _, err := ctx.PageDict(context.Background(), p, false)
	if err != nil {
		return false, err
	}
	o, ok := d.Find("Tabs")
	if !ok || o == nil {
		return false, nil
	}
	o, err = ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	if o == nil {
		return false, nil
	}
	n, ok := o.(types.Name)
	if !ok {
		return false, errBadTabs
	}
	return n.Value() == "S", nil
}
