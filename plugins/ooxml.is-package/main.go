// Command ooxml.is-package answers: is the file an Office Open XML package
// with a main document part?
package main

import (
	"os"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, nil
		}
		root := p.Part("_rels/.rels")
		if root == nil {
			return false, nil
		}
		rels, err := ooxml.Rels(root)
		if err != nil {
			return false, nil
		}
		for _, r := range rels {
			if r.IsType("officeDocument") && !strings.EqualFold(r.TargetMode, "External") &&
				p.Part(ooxml.RootTarget(r.Target)) != nil {
				return true, nil
			}
		}
		return false, nil
	})
}
