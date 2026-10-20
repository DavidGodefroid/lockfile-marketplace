// Command pdf.leaves-metadata-unencrypted answers: does the encrypted PDF
// leave its metadata unencrypted?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		st, err := f.Stat()
		if err != nil {
			return false, err
		}
		// Read raw: pdfx.Open cannot open a file that needs a user password.
		d, err := pdfx.EncryptDict(f, st.Size())
		if err != nil || d == nil {
			return false, err
		}
		v, ok := d["EncryptMetadata"]
		if !ok || v == nil {
			return false, nil
		}
		b, ok := v.(types.Boolean)
		if !ok {
			return false, pdfx.ErrBadEncrypt
		}
		return !b.Value(), nil
	})
}
