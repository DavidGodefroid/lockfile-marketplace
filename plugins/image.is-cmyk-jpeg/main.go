// Command image.is-cmyk-jpeg answers: is the file a CMYK JPEG?
package main

import (
	"fmt"
	"image/color"
	"image/jpeg"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return isCMYK(f)
	})
}

// isCMYK reports whether r is a JPEG whose colour model, as read from its
// headers by jpeg.DecodeConfig, is CMYK. No pixel is decoded. Input that is
// not a JPEG, or whose headers do not parse, is an error.
func isCMYK(r io.Reader) (bool, error) {
	cfg, err := jpeg.DecodeConfig(r)
	if err != nil {
		return false, fmt.Errorf("not a readable JPEG: %w", err)
	}
	return cfg.ColorModel == color.CMYKModel, nil
}
