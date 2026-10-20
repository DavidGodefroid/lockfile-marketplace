// Command pdf.uses-rc4-encryption answers: is the PDF encrypted with RC4?
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
		d, err := pdfx.EncryptDict(f, st.Size())
		if err != nil || d == nil {
			return false, err
		}
		return usesRC4(d)
	})
}

// usesRC4 reports whether /V is below 4 (RC4 by the security handler's
// key length; absent means 0, PDF 32000-1 table 20) or a crypt filter has
// /CFM /V2 (RC4).
func usesRC4(d types.Dict) (bool, error) {
	v := 0
	if o, ok := d["V"]; ok {
		n, ok := o.(types.Integer)
		if !ok {
			return false, pdfx.ErrBadEncrypt
		}
		v = n.Value()
	}
	methods, err := pdfx.CryptFilterMethods(d)
	if err != nil {
		return false, err
	}
	if v < 4 {
		return true, nil
	}
	for _, m := range methods {
		if m == "V2" {
			return true, nil
		}
	}
	return false, nil
}
