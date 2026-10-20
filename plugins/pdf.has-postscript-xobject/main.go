// Command pdf.has-postscript-xobject answers: does the PDF hold a PostScript
// XObject?
//
// Every stream of the file is visited, orphans included. A stream counts when
// its /Subtype is /PS, or when it is a form XObject (/Subtype /Form) whose
// /Subtype2 is /PS. Stream data is never decoded. The file is read with
// pdfx.Read, not pdfx.Open: pdfcpu's validation rejects a PostScript XObject
// drawn from a page, which would make the main positive case an error.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// name returns the name stored under key in d, dereferenced; "" when the key
// is absent or not a name.
func name(ctx *model.Context, d types.Dict, key string) (types.Name, error) {
	o, ok := d.Find(key)
	if !ok {
		return "", nil
	}
	o, err := ctx.Dereference(o)
	if err != nil {
		return "", err
	}
	n, _ := o.(types.Name)
	return n, nil
}

// postScript reports whether the stream dict d is a PostScript XObject.
func postScript(ctx *model.Context, d types.Dict) (bool, error) {
	sub, err := name(ctx, d, "Subtype")
	if err != nil {
		return false, err
	}
	switch sub {
	case "PS":
		return true, nil
	case "Form":
		sub2, err := name(ctx, d, "Subtype2")
		return sub2 == "PS", err
	}
	return false, nil
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Read(f)
		if err != nil {
			return false, err
		}
		found := false
		var perr error
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			s, ok := o.(types.StreamDict)
			if !ok {
				return false
			}
			found, perr = postScript(ctx, s.Dict)
			return found || perr != nil
		})
		if err != nil {
			return false, err
		}
		return found, perr
	})
}
