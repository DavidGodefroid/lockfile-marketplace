// Command pdf.has-unsigned-signature-field answers: does the PDF have an
// unsigned signature field?
//
// A signature field is a terminal node (no /Kids) of the AcroForm field
// tree whose field type, inherited from its ancestors, is Sig. It is
// unsigned when neither it nor any ancestor (/Parent) has a non-null /V.
// The value itself is never read or verified.
package main

import (
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// maxParents bounds the /Parent walk, as pdfx bounds the field tree.
const maxParents = 64

var errParentChainTooLong = errors.New("pdf.has-unsigned-signature-field: /Parent chain too long")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		var verr error
		err = pdfx.EachField(ctx, func(field types.Dict, ft string) bool {
			if ft != "Sig" {
				return false
			}
			if _, ok := field.Find("Kids"); ok {
				return false
			}
			signed, e := hasValue(ctx, field)
			if e != nil {
				verr = e
				return true
			}
			found = !signed
			return found
		})
		if err == nil {
			err = verr
		}
		return found, err
	})
}

// hasValue reports whether field or one of its ancestors has a non-null /V.
func hasValue(ctx *model.Context, field types.Dict) (bool, error) {
	d := field
	for range maxParents {
		if o, ok := d.Find("V"); ok && o != nil {
			return true, nil
		}
		p, ok := d.Find("Parent")
		if !ok || p == nil {
			return false, nil
		}
		parent, err := ctx.DereferenceDict(p)
		if err != nil {
			return false, err
		}
		if parent == nil {
			return false, nil
		}
		d = parent
	}
	return false, errParentChainTooLong
}
