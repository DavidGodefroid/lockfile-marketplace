// Command pdf.has-xref-stream answers: does the PDF use cross-reference
// streams? Readers limited to PDF 1.4 cannot open such a file; a hybrid
// file stays readable to them and does not count.
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
		for _, s := range chain {
			if s.Stream {
				return true, nil
			}
		}
		return false, nil
	})
}
