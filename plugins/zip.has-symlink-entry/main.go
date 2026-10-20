// Command zip.has-symlink-entry answers: does the ZIP archive have a
// symbolic-link entry?
package main

import (
	"io/fs"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		// Mode reads the Unix mode from the external attributes only when the
		// creator is Unix-like; MS-DOS attributes cannot express a symlink.
		for _, e := range r.File {
			if e.Mode()&fs.ModeSymlink != 0 {
				return true, nil
			}
		}
		return false, nil
	})
}
