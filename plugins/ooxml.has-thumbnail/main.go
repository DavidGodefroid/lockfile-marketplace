// Command ooxml.has-thumbnail answers: does the Office Open XML package
// carry a thumbnail?
package main

import (
	"os"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// relThumbnail is the OPC thumbnail relationship type. It lives under the
// package relationships base, which Rel.IsType does not cover, and strict
// packages keep it unchanged.
const relThumbnail = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail"

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		root := p.Part("_rels/.rels")
		if root == nil {
			return false, nil
		}
		rels, err := ooxml.Rels(root)
		if err != nil {
			return false, err
		}
		for _, r := range rels {
			if r.Type != relThumbnail && !r.IsType("metadata/thumbnail") {
				continue
			}
			// An external target or a dangling one is no thumbnail in
			// the package.
			if strings.EqualFold(r.TargetMode, "External") {
				continue
			}
			if e := p.Part(ooxml.RootTarget(r.Target)); e != nil && !strings.HasSuffix(e.Name, "/") {
				return true, nil
			}
		}
		return false, nil
	})
}
