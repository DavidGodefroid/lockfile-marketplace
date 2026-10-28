// Command pdf.has-redaction-annotation answers: does some page carry a
// redaction annotation (content marked for redaction but not yet removed)?
//
// Only annotations a page lists in /Annots count. A /Redact annotation
// (ISO 32000-2 §12.5.6.23) marks a region; once the redaction is applied
// the annotation is removed, so its presence means it was not applied.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		err = pdfx.EachAnnotation(ctx, func(_ int, a types.Dict) bool {
			s := a.NameEntry("Subtype")
			found = s != nil && *s == "Redact"
			return found
		})
		return found, err
	})
}
