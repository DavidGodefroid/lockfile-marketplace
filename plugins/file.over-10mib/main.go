// Command file.over-10mib answers: is the file larger than 10 MiB?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// limit is 10 MiB; a file of exactly this size answers 0.
const limit = 10 << 20

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		fi, err := f.Stat()
		if err != nil {
			return false, err
		}
		return fi.Size() > limit, nil
	})
}
