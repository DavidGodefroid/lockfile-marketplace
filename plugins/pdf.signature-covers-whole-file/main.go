// Command pdf.signature-covers-whole-file answers: does a signature cover
// the whole file?
//
// A signature covers the whole file when its /ByteRange is [0 b c d] with
// b < c and c+d equal to the file size: no byte was appended after
// signing (no incremental update, no trailing data). Nothing is verified:
// neither the cryptography nor the certificate.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		st, err := f.Stat()
		if err != nil {
			return false, err
		}
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		sigs, err := pdfx.SignatureValues(ctx)
		if err != nil {
			return false, err
		}
		for _, sig := range sigs {
			covers, err := pdfx.CoversWholeFile(ctx, sig, st.Size())
			if err != nil || covers {
				return covers, err
			}
		}
		return false, nil
	})
}
