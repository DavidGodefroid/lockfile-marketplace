// Command pdf.has-incremental-update answers: was the PDF modified by at
// least one incremental update after its first save?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		// A file that does not parse gets no structural answer.
		if _, err := pdfx.Open(f); err != nil {
			return false, err
		}
		fi, err := f.Stat()
		if err != nil {
			return false, err
		}
		chain, err := pdfx.XRefChain(f, fi.Size())
		if err != nil {
			return false, err
		}
		lin, err := pdfx.IsLinearized(f, fi.Size())
		if err != nil {
			return false, err
		}
		return pdfx.CountRevisions(chain, lin) > 1, nil
	})
}
