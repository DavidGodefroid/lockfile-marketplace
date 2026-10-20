// Command pdf.has-embedded-file answers: does the PDF embed a file?
//
// Every object of the file is visited, orphans included. The embedded
// file, its name and its description are never read.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func isName(d types.Dict, key, value string) bool {
	n := d.NameEntry(key)
	return n != nil && *n == value
}

// embeds reports whether o is a file specification with an /EF entry, an
// embedded file stream or a file attachment annotation.
func embeds(o types.Object) bool {
	d, ok := o.(types.Dict)
	if !ok {
		return false
	}
	if _, ok := d.Find("EF"); ok {
		return true
	}
	return isName(d, "Type", "EmbeddedFile") || isName(d, "Subtype", "FileAttachment")
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		cat, err := ctx.Catalog()
		if err != nil {
			return false, err
		}
		if o, ok := cat.Find("Names"); ok {
			names, err := ctx.DereferenceDict(o)
			if err != nil {
				return false, err
			}
			if _, ok := names.Find("EmbeddedFiles"); ok {
				return true, nil
			}
		}
		found := false
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			found = embeds(o)
			return found
		})
		return found, err
	})
}
