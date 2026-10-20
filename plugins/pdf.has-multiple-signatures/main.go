// Command pdf.has-multiple-signatures answers: does the PDF carry two or
// more digital signatures?
//
// Signatures are counted as pdf.has-signature finds them: the values of
// the signature fields of the AcroForm field tree, document time stamps
// included. Nothing is verified and the signers are never read.
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
		return len(sigs) >= 2, err
	})
}
