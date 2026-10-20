// Command pdf.printing-restricted answers: does the PDF forbid printing?
//
// The permissions are the /P entry of the /Encrypt dictionary. A PDF that
// needs a user password is an error (pdf.needs-password answers that).
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, encrypted, err := pdfx.Permissions(f)
		if err != nil || !encrypted {
			return false, err
		}
		return p&pdfx.PermPrint == 0, nil
	})
}
