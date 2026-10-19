// Command pdf.truncated answers: does the PDF look truncated?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		fi, err := f.Stat()
		if err != nil {
			return false, err
		}
		return pdfx.LooksTruncated(f, fi.Size())
	})
}
