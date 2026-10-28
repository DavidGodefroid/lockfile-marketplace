// Command pdf.claims-facturx answers: does the PDF declare Factur-X /
// ZUGFeRD conformance in its XMP metadata? It tests only for the
// ConformanceLevel property, by namespace and name; no value of the
// metadata is read or kept.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// facturXNamespace is the Factur-X / ZUGFeRD 2 PDF/A extension schema.
const facturXNamespace = "urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#"

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
		return pdfx.HasXMPProperty(utf8, facturXNamespace, "ConformanceLevel")
	})
}
