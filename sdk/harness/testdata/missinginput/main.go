// Its expected.toml leaves an input without a value: Verify refuses it
// before running it.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(*os.File) (bool, error) {
		_, err := sdk.Inputs()
		return true, err
	}, sdk.WithDetail())
}
