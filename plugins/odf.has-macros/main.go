// Command odf.has-macros answers: does the OpenDocument package contain
// macros?
package main

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// mimetypeLen bounds the read of the mimetype entry; every ODF media
// type is far shorter, so a longer entry is not an ODF mimetype.
const mimetypeLen = 128

// odfMediaPrefix starts the media type of every OpenDocument package.
const odfMediaPrefix = "application/vnd.oasis.opendocument."

// macroDirs are the package directories holding Basic libraries and
// scripts in other languages.
var macroDirs = []string{"Basic/", "Scripts/"}

var errNotODF = errors.New("not an OpenDocument package: the first entry is not an ODF mimetype")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		if len(r.File) == 0 || r.File[0].Name != "mimetype" {
			return false, errNotODF
		}
		mt, err := zipx.Prefix(r.File[0], mimetypeLen)
		if err != nil {
			return false, err
		}
		if len(mt) == mimetypeLen || len(mt) <= len(odfMediaPrefix) || !bytes.HasPrefix(mt, []byte(odfMediaPrefix)) {
			return false, errNotODF
		}
		for _, e := range r.File[1:] {
			for _, d := range macroDirs {
				if strings.HasPrefix(e.Name, d) && len(e.Name) > len(d) {
					return true, nil
				}
			}
		}
		return false, nil
	})
}
