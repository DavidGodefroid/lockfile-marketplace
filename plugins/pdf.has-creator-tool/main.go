// Command pdf.has-creator-tool answers: does the PDF name the tool that created it in its
// metadata? The tool name is never returned or kept: only whether it is empty.
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// maxXMP caps the decoded catalog metadata.
const maxXMP = 8 << 20

// xmpNamespace is the XMP basic schema namespace, home of xmp:CreatorTool.
const xmpNamespace = "http://ns.adobe.com/xap/1.0/"

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		ctx, err := pdfx.Open(f)
		if err != nil {
			return false, err
		}
		ok, err := pdfx.InfoHasNonEmpty(ctx, "Creator")
		if err != nil || ok {
			return ok, err
		}
		xmp, ok, err := pdfx.CatalogXMP(ctx, maxXMP)
		if err != nil || !ok {
			return false, err
		}
		utf8, err := pdfx.ToUTF8(xmp)
		if err != nil {
			return false, err
		}
		return pdfx.HasXMPProperty(utf8, xmpNamespace, "CreatorTool")
	})
}
