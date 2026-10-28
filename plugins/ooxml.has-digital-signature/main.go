// Command ooxml.has-digital-signature answers: is the Office Open XML
// package digitally signed (OPC package signature)?
package main

import (
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// OPC digital-signature relationship types (ECMA-376 Part 2). They live
// under the package relationships base, which strict OOXML keeps, so
// Rel.IsType (officeDocument bases) does not apply and the match is exact.
const (
	relOrigin    = "http://schemas.openxmlformats.org/package/2006/relationships/digital-signature/origin"
	relSignature = "http://schemas.openxmlformats.org/package/2006/relationships/digital-signature/signature"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		origins, err := targets(p, "", relOrigin)
		if err != nil {
			return false, err
		}
		for _, origin := range origins {
			sigs, err := targets(p, origin, relSignature)
			if err != nil {
				return false, err
			}
			if len(sigs) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
}

// targets returns the existing parts that the relationships of type typ
// of source (the package root when "") point to. A missing .rels part
// means no relationships; a malformed one is an error.
func targets(p *ooxml.Package, source, typ string) ([]string, error) {
	dir, relsName := "", "_rels/.rels"
	if source != "" {
		dir = path.Dir(source)
		relsName = path.Join(dir, "_rels", path.Base(source)+".rels")
	}
	e := p.Part(relsName)
	if e == nil {
		return nil, nil
	}
	rels, err := ooxml.Rels(e)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, r := range rels {
		if r.Type != typ || r.TargetMode == "External" {
			continue
		}
		name := r.Target
		if !strings.HasPrefix(name, "/") {
			name = path.Join(dir, name)
		}
		name = ooxml.RootTarget(name)
		if p.Part(name) != nil {
			out = append(out, name)
		}
	}
	return out, nil
}
