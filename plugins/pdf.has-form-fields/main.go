// Command pdf.has-form-fields answers: does the PDF have fillable form
// fields?
//
// A form field is a node of the AcroForm field tree whose field type,
// inherited from its ancestors, is a button (Btn), a text field (Tx) or a
// choice (Ch). Signature fields (pdf.has-signature) and XFA forms
// (pdf.has-xfa) do not count; field values are never read.
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
		err = pdfx.EachField(ctx, func(_ types.Dict, ft string) bool {
			found = ft == "Btn" || ft == "Tx" || ft == "Ch"
			return found
		})
		return found, err
	})
}
