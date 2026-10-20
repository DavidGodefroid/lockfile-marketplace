// Command pdf.has-xmp-metadata answers: does the PDF have document-level XMP metadata? The
// metadata stream is neither decoded nor parsed.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		return pdfx.HasCatalogMetadata(ctx)
	})
}
