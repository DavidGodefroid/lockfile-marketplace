// Command pdf.has-cmyk-image answers: does the PDF have a CMYK image?
//
// Every object of the file is visited, orphans included. Only the /Subtype
// and /ColorSpace entries of image XObjects are read; image data is never
// decoded.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

var (
	errColorSpace = errors.New("pdf.has-cmyk-image: image /ColorSpace is neither a name nor an array")
	errICC        = errors.New("pdf.has-cmyk-image: /ICCBased colour space without a stream carrying an integer /N")
)

// isImage reports whether the stream dict d is an image XObject.
func isImage(ctx *model.Context, d types.Dict) (bool, error) {
	o, ok := d.Find("Subtype")
	if !ok {
		return false, nil
	}
	o, err := ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	n, ok := o.(types.Name)
	return ok && n == "Image", nil
}

// isCMYK reports whether the image dict d has a /ColorSpace that is
// /DeviceCMYK, or an /ICCBased profile with four components.
func isCMYK(ctx *model.Context, d types.Dict) (bool, error) {
	o, ok := d.Find("ColorSpace")
	if !ok {
		return false, nil
	}
	o, err := ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	switch v := o.(type) {
	case nil:
		return false, nil
	case types.Name:
		return v == "DeviceCMYK", nil
	case types.Array:
		if len(v) == 0 {
			return false, errColorSpace
		}
		family, err := ctx.Dereference(v[0])
		if err != nil {
			return false, err
		}
		if n, ok := family.(types.Name); !ok || n != "ICCBased" {
			return false, nil
		}
		if len(v) < 2 {
			return false, errICC
		}
		profile, err := ctx.Dereference(v[1])
		if err != nil {
			return false, err
		}
		sd, ok := profile.(types.StreamDict)
		if !ok {
			return false, errICC
		}
		n, err := ctx.Dereference(sd.Dict["N"])
		if err != nil {
			return false, err
		}
		i, ok := n.(types.Integer)
		if !ok {
			return false, errICC
		}
		return i == 4, nil
	}
	return false, errColorSpace
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		var ferr error
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			sd, ok := o.(types.StreamDict)
			if !ok {
				return false
			}
			img, err := isImage(ctx, sd.Dict)
			if err != nil || !img {
				ferr = err
				return err != nil
			}
			found, ferr = isCMYK(ctx, sd.Dict)
			return found || ferr != nil
		})
		if err == nil {
			err = ferr
		}
		return found, err
	})
}
