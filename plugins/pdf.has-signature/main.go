// Command pdf.has-signature answers: is the PDF digitally signed?
//
// A signature is the value of a signature field of the AcroForm field
// tree. Nothing is verified: neither the cryptography, the certificate nor
// what the signature covers (pdf.signature-covers-whole-file); the signer
// is never read.
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
		return len(sigs) > 0, err
	})
}
