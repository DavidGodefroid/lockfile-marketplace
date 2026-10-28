// Command pdf.has-pades-signature answers: does the PDF carry a PAdES
// signature?
//
// Signatures are found as pdf.has-signature finds them: the values of the
// signature fields of the AcroForm field tree. A value is PAdES when its
// /SubFilter is /ETSI.CAdES.detached. Nothing is verified and the signers
// are never read.
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
		sigs, err := pdfx.SignatureValues(ctx)
		if err != nil {
			return false, err
		}
		for _, v := range sigs {
			if s := v.NameEntry("SubFilter"); s != nil && *s == "ETSI.CAdES.detached" {
				return true, nil
			}
		}
		return false, nil
	})
}
