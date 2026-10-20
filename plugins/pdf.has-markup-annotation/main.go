// Command pdf.has-markup-annotation answers: does some page carry a markup
// annotation (a comment, highlight, stamp, drawing…)?
//
// Only annotations a page lists in /Annots count. Markup is the subtype
// list of ISO 32000-2 §12.5.6.2; Link, Widget, Popup and the other
// non-markup subtypes do not count.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var markup = map[string]bool{
	"Text": true, "FreeText": true, "Line": true, "Square": true, "Circle": true,
	"Polygon": true, "PolyLine": true, "Highlight": true, "Underline": true,
	"Squiggly": true, "StrikeOut": true, "Stamp": true, "Caret": true, "Ink": true,
	"FileAttachment": true, "Sound": true, "Redact": true, "Projection": true,
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachAnnotation(ctx, func(_ int, a types.Dict) bool {
			s := a.NameEntry("Subtype")
			found = s != nil && markup[*s]
			return found
		})
		return found, err
	})
}
