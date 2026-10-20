// Command pdf.has-multimedia-annotation answers: does some page carry a
// multimedia annotation (embedded video, sound or rich media)?
//
// Only annotations a page lists in /Annots count. Multimedia is the
// Screen, Movie, Sound and RichMedia subtypes; every other subtype,
// including a Link to a media file, does not count.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var multimedia = map[string]bool{
	"Screen": true, "Movie": true, "Sound": true, "RichMedia": true,
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
			found = s != nil && multimedia[*s]
			return found
		})
		return found, err
	})
}
