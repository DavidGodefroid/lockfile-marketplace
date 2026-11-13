// Answers like test.contains but echoes the typed value in its detail,
// which Verify refuses.
package main

import (
	"bytes"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		in, err := sdk.Inputs()
		if err != nil {
			return false, err
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return false, err
		}
		io.WriteString(sdk.Detail(), "looked for "+in["needle"])
		return bytes.Contains(b, []byte(in["needle"])), nil
	}, sdk.WithDetail())
}
