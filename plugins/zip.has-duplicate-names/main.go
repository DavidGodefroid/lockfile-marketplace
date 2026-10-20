// Command zip.has-duplicate-names answers: does the ZIP archive have two
// entries with the same name?
package main

import (
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
		// Names compare as exact bytes: no case folding, no path cleaning.
		seen := make(map[string]struct{}, len(r.File))
		for _, e := range r.File {
			if _, ok := seen[e.Name]; ok {
				return true, nil
			}
			seen[e.Name] = struct{}{}
		}
		return false, nil
	})
}
