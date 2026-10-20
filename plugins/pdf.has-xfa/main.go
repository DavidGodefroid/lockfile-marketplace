// Command pdf.has-xfa answers: does the PDF carry an XFA form?
//
// An XFA form is the /XFA entry of the catalog's interactive form
// dictionary, a stream or an array of packets; its XML is never read.
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
		form, err := pdfx.AcroForm(ctx)
		if err != nil || form == nil {
			return false, err
		}
		_, ok := form.Find("XFA")
		return ok, nil
	})
}
