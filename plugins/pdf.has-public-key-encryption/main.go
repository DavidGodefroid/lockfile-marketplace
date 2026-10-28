// Command pdf.has-public-key-encryption answers: is the PDF encrypted with a public-key security handler?
//
// The handler is the /Filter of the trailer's /Encrypt dictionary, read raw
// (pdfx.EncryptDict): pdfx.Open cannot open a file encrypted for recipient
// certificates. A /Filter that is absent or not a name is an error.
package main

import (
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var errBadFilter = errors.New("/Encrypt /Filter is absent or not a name")

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
		filter, ok := d["Filter"].(types.Name)
		if !ok {
			return false, errBadFilter
		}
		return filter != "Standard", nil
	})
}
