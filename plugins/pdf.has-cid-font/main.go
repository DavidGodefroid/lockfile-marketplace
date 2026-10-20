// Command pdf.has-cid-font answers: does the PDF use a CID-keyed (composite)
// font?
//
// Every object of the file is visited, so fonts in object streams and fonts
// no page references count. Only the font dictionary's /Subtype is read: the
// font name, encoding and glyphs never are.
package main

import (
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// isType0Font reports whether o is a composite font dictionary: /Subtype
// /Type0, with /Type /Font when /Type is present.
func isType0Font(o types.Object) bool {
	d, ok := o.(types.Dict)
	if !ok {
		return false
	}
	if s := d.NameEntry("Subtype"); s == nil || *s != "Type0" {
		return false
	}
	t := d.NameEntry("Type")
	return t == nil || *t == "Font"
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			found = isType0Font(o)
			return found
		})
		return found, err
	})
}
