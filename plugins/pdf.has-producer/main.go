// Command pdf.has-producer answers: does the PDF name the software that produced it in its
// metadata? The producer is never returned or kept: only whether it is empty.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxXMP caps the decoded catalog metadata.
const maxXMP = 8 << 20

// pdfNamespace is the XMP namespace of the Adobe PDF schema, which holds pdf:Producer.
const pdfNamespace = "http://ns.adobe.com/pdf/1.3/"

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		ok, err := pdfx.InfoHasNonEmpty(ctx, "Producer")
		if err != nil || ok {
			return ok, err
		}
		xmp, ok, err := pdfx.CatalogXMP(ctx, maxXMP)
		if err != nil || !ok {
			return false, err
		}
		utf8, err := pdfx.ToUTF8(xmp)
		if err != nil {
			return false, err
		}
		return pdfx.HasXMPProperty(utf8, pdfNamespace, "Producer")
	})
}
