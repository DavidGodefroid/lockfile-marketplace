// Command pdf.has-document-timestamp answers: does the PDF carry a
// document time stamp?
//
// A document time stamp is a signature value (pdfx.SignatureValues) whose
// /Type is /DocTimeStamp or whose /SubFilter is /ETSI.RFC3161 (ISO 32000-2
// 12.8.5). Nothing is verified: neither the time stamp token nor the
// authority; the time is never read.
package main

import (
	"fmt"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		sigs, err := pdfx.SignatureValues(ctx)
		if err != nil {
			return false, err
		}
		for _, sig := range sigs {
			ts, err := isTimestamp(ctx, sig)
			if err != nil || ts {
				return ts, err
			}
		}
		return false, nil
	})
}

// isTimestamp reports whether sig is a document time stamp. A /Type or
// /SubFilter that is not a name is an error.
func isTimestamp(ctx *model.Context, sig types.Dict) (bool, error) {
	for _, want := range [][2]string{{"Type", "DocTimeStamp"}, {"SubFilter", "ETSI.RFC3161"}} {
		o, ok := sig.Find(want[0])
		if !ok || o == nil {
			continue
		}
		o, err := ctx.Dereference(o)
		if err != nil {
			return false, err
		}
		n, ok := o.(types.Name)
		if !ok {
			return false, fmt.Errorf("signature /%s is not a name", want[0])
		}
		if n.Value() == want[1] {
			return true, nil
		}
	}
	return false, nil
}
