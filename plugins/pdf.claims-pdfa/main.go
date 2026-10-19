// Command pdf.claims-pdfa answers: does the PDF declare PDF/A conformance
// in its XMP metadata? It tests only for the pdfaid:part property, by
// namespace and name; no value of the metadata is read or kept.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxXMP caps the decoded catalog metadata.
const maxXMP = 8 << 20

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		xmp, ok, err := pdfx.CatalogXMP(ctx, maxXMP)
		if err != nil || !ok {
			return false, err
		}
		utf8, err := pdfx.ToUTF8(xmp)
		if err != nil {
			return false, err
		}
		return pdfx.HasXMPProperty(utf8, pdfx.PDFAIDNamespace, "part")
	})
}
