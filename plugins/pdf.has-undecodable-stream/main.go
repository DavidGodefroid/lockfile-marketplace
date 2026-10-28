// Command pdf.has-undecodable-stream answers: does the PDF have a stream
// whose data fails to decode with its declared filters?
//
// Every stream of the file is decoded, orphans included, one at a time
// and bounded; the decoded data is never read.
package main

import (
	"errors"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/filter"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxDecoded bounds the decoded size of one stream; a larger stream is an
// error, not an answer.
const maxDecoded = 64 << 20

// imageFilter reports whether the pipeline has an image codec, whose data
// is not decoded.
func imageFilter(sd types.StreamDict) bool {
	for _, f := range sd.FilterPipeline {
		switch f.Name {
		case filter.DCT, filter.JPX, filter.JBIG2, filter.CCITTFax:
			return true
		}
	}
	return false
}

// undecodable decodes sd (a copy: the context is left as it is).
func undecodable(sd types.StreamDict) (bool, error) {
	// pdfcpu takes empty data as decoded to nothing.
	if len(sd.Raw) == 0 || imageFilter(sd) {
		return false, nil
	}
	err := sd.DecodeWithLimit(maxDecoded)
	if errors.Is(err, filter.ErrDecodeLimitExceeded) {
		return false, err
	}
	return err != nil, nil
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		found := false
		var decodeErr error
		err = pdfx.EachObject(ctx, func(o types.Object) bool {
			sd, ok := o.(types.StreamDict)
			if !ok {
				return false
			}
			found, decodeErr = undecodable(sd)
			return found || decodeErr != nil
		})
		if err != nil {
			return false, err
		}
		return found, decodeErr
	})
}
