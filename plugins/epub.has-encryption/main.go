// Command epub.has-encryption answers: does the EPUB declare encrypted
// resources (DRM or font obfuscation)?
package main

import (
	"archive/zip"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// mimetype is the OCF media type the first entry, named "mimetype", holds.
const mimetype = "application/epub+zip"

// encryptionPath is the OCF container file listing encrypted resources.
const encryptionPath = "META-INF/encryption.xml"

var errNotEPUB = errors.New("not an EPUB: first entry is not mimetype application/epub+zip")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		if err := checkMimetype(r); err != nil {
			return false, err
		}
		for _, e := range r.File {
			if e.Name == encryptionPath && !e.FileInfo().IsDir() {
				return true, nil
			}
		}
		return false, nil
	})
}

// checkMimetype requires the first entry to be "mimetype" holding exactly
// the EPUB media type.
func checkMimetype(r *zip.Reader) error {
	if len(r.File) == 0 || r.File[0].Name != "mimetype" {
		return errNotEPUB
	}
	p, err := zipx.Prefix(r.File[0], len(mimetype)+1)
	if err != nil {
		return err
	}
	if string(p) != mimetype {
		return errNotEPUB
	}
	return nil
}
