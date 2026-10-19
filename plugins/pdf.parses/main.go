// Command pdf.parses answers: does the file parse as a PDF without errors?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		_, err := pdfx.Open(f)
		return err == nil, nil
	})
}
