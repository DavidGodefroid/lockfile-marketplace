// Command pdf.has-font-without-tounicode answers: does the PDF have a font
// whose text cannot be mapped to Unicode?
//
// Every font dict of the file is visited, used or not. A Type0 font needs a
// /ToUnicode map. A simple font (Type1, MMType1, TrueType, Type3) needs a
// /ToUnicode map or a standard encoding: /Encoding naming WinAnsiEncoding,
// MacRomanEncoding or MacExpertEncoding, or an encoding dict whose
// /BaseEncoding names one of them. CIDFont descendants are judged through
// their Type0 parent.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

var errEncoding = errors.New("font /Encoding is neither a name nor a dict")

var simple = map[string]bool{"Type1": true, "MMType1": true, "TrueType": true, "Type3": true}

var standard = map[string]bool{"WinAnsiEncoding": true, "MacRomanEncoding": true, "MacExpertEncoding": true}

// lacksUnicode reports whether o is a font dict whose text has no Unicode
// mapping.
func lacksUnicode(ctx *model.Context, o types.Object) (bool, error) {
	d, ok := o.(types.Dict)
	if !ok {
		return false, nil
	}
	if t := d.NameEntry("Type"); t == nil || *t != "Font" {
		return false, nil
	}
	sub := d.NameEntry("Subtype")
	if sub == nil || (*sub != "Type0" && !simple[*sub]) {
		return false, nil
	}
	if u, ok := d.Find("ToUnicode"); ok && u != nil {
		return false, nil
	}
	if *sub == "Type0" {
		return true, nil
	}
	std, err := standardEncoding(ctx, d)
	return !std, err
}

// standardEncoding reports whether the simple font d has a standard
// /Encoding.
func standardEncoding(ctx *model.Context, d types.Dict) (bool, error) {
	o, ok := d.Find("Encoding")
	if !ok {
		return false, nil
	}
	enc, err := ctx.Dereference(o)
	if err != nil {
		return false, err
	}
	switch v := enc.(type) {
	case nil:
		return false, nil
	case types.Name:
		return standard[v.Value()], nil
	case types.Dict:
		b := v.NameEntry("BaseEncoding")
		return b != nil && standard[*b], nil
	}
	return false, errEncoding
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
			found, ferr = lacksUnicode(ctx, o)
			return found || ferr != nil
		})
		if err != nil {
			return false, err
		}
		return found, ferr
	})
}
