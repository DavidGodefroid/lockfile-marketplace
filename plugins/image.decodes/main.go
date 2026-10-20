// Command image.decodes answers: does the file decode as a PNG or JPEG image?
package main

import (
	"fmt"
	"image"
	_ "image/jpeg" // registers the only two decoders image.Decode may use
	_ "image/png"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxPixels caps the declared image size before the pixels are decoded, so a
// small file that declares a huge image cannot exhaust memory.
const maxPixels = 32_000_000

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return decodes(f)
	})
}

// decodes reports whether r holds a PNG or JPEG image that decodes without
// error. A header that does not parse is a 0; a header that declares more
// than maxPixels pixels is an error, not a verdict.
func decodes(r io.ReadSeeker) (bool, error) {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return false, nil
	}
	if px := int64(cfg.Width) * int64(cfg.Height); px > maxPixels {
		return false, fmt.Errorf("image is %d×%d pixels, over the %d-pixel limit", cfg.Width, cfg.Height, maxPixels)
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	_, _, err = image.Decode(r)
	return err == nil, nil
}
