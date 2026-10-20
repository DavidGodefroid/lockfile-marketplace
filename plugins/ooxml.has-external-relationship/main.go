// Command ooxml.has-external-relationship answers: does the Office Open
// XML package have a relationship to an external target?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// external is the OPC TargetMode of a target outside the package. The
// schema enumeration is case-sensitive, so the match is exact.
const external = "External"

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		// Every .rels part is parsed before answering 0, so a malformed
		// one is an error rather than a silent "no".
		for _, e := range p.RelsParts() {
			rels, err := ooxml.Rels(e)
			if err != nil {
				return false, err
			}
			for _, r := range rels {
				if r.TargetMode == external {
					return true, nil
				}
			}
		}
		return false, nil
	})
}
