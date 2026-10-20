// Command pdf.has-named-destinations answers: does the PDF define named destinations? Only the
// presence of an entry in the catalog /Dests dictionary or in the /Names /Dests name tree is
// tested; no name or target is read.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return hasNamedDestinations(ctx)
	})
}

// hasNamedDestinations reports whether the PDF 1.1 catalog /Dests dictionary
// has an entry, or the PDF 1.2 /Names /Dests name tree has a leaf or a kid.
func hasNamedDestinations(ctx *model.Context) (bool, error) {
	dests, err := pdfx.CatalogDict(ctx, "Dests")
	if err != nil {
		return false, err
	}
	if len(dests) > 0 {
		return true, nil
	}
	names, err := pdfx.CatalogDict(ctx, "Names")
	if err != nil || names == nil {
		return false, err
	}
	o, found := names.Find("Dests")
	if !found || o == nil {
		return false, nil
	}
	tree, err := ctx.DereferenceDict(o)
	if err != nil {
		return false, pdfx.ErrWrongType
	}
	for _, key := range []string{"Names", "Kids"} {
		ok, err := nonEmptyArray(ctx, tree, key)
		if err != nil || ok {
			return ok, err
		}
	}
	return false, nil
}

// nonEmptyArray reports whether d[key] dereferences to an array with at least
// one element; a value of another type is an error.
func nonEmptyArray(ctx *model.Context, d types.Dict, key string) (bool, error) {
	o, found := d.Find(key)
	if !found || o == nil {
		return false, nil
	}
	a, err := ctx.DereferenceArray(o)
	if err != nil {
		return false, pdfx.ErrWrongType
	}
	return len(a) > 0, nil
}
