// Command pdf.has-piece-info answers: does the document carry private
// application data (/PieceInfo on the catalog or on a page)?
//
// Only the presence of a /PieceInfo dictionary counts; its content is
// never read. A /PieceInfo that is not a dictionary is an error, a null
// one counts as absent. Pages are those of the page tree.
package main

import (
	"context"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return hasPieceInfo(ctx)
	})
}

func hasPieceInfo(ctx *model.Context) (bool, error) {
	if ctx.PageCount == 0 {
		return false, pdfx.ErrNoPages
	}
	d, err := pdfx.CatalogDict(ctx, "PieceInfo")
	if err != nil || d != nil {
		return d != nil, err
	}
	for p := 1; p <= ctx.PageCount; p++ {
		page, _, _, err := ctx.PageDict(context.Background(), p, false)
		if err != nil {
			return false, err
		}
		o, ok := page.Find("PieceInfo")
		if !ok {
			continue
		}
		d, err := ctx.DereferenceDict(o)
		if err != nil {
			return false, err
		}
		if d != nil {
			return true, nil
		}
	}
	return false, nil
}
