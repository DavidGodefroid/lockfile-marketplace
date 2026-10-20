// Command pdf.is-linearized answers: is the PDF linearized for fast web
// view? A linearization dictionary whose /L no longer matches the file size
// (an update was appended) does not count.
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
		ok, length, err := pdfx.Linearization(f, fi.Size())
		if err != nil || !ok {
			return false, err
		}
		return length == fi.Size(), nil
	})
}
