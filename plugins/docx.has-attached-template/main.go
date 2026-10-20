// Command docx.has-attached-template answers: is the Word document attached
// to a template?
package main

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// errNoDocument is returned for a package with no word/document.xml.
var errNoDocument = errors.New("docx: no word/document.xml")

const document = "word/document.xml"

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		if p.Part(document) == nil {
			return false, errNoDocument
		}
		// The settings part is the target of the main document's settings
		// relationship.
		settings, err := targets(p, document, "settings")
		if err != nil {
			return false, err
		}
		for _, s := range settings {
			if p.Part(s) == nil {
				continue
			}
			t, err := targets(p, s, "attachedTemplate")
			if err != nil {
				return false, err
			}
			if len(t) > 0 {
				return true, nil
			}
		}
		return false, nil
	})
}

// targets returns the targets of the relationships of type short held by
// part's .rels part, resolved to part names for internal targets. A part
// with no .rels part has none.
func targets(p *ooxml.Package, part, short string) ([]string, error) {
	dir, base := path.Split(part)
	e := p.Part(dir + "_rels/" + base + ".rels")
	if e == nil {
		return nil, nil
	}
	rels, err := ooxml.Rels(e)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, r := range rels {
		if !r.IsType(short) {
			continue
		}
		// An external target (the usual case for a template) is kept as
		// it is; an internal one is relative to the source part.
		t := r.Target
		if !strings.EqualFold(r.TargetMode, "External") {
			if strings.HasPrefix(t, "/") {
				t = ooxml.RootTarget(t)
			} else {
				t = ooxml.RootTarget(path.Join(dir, t))
			}
		}
		out = append(out, t)
	}
	return out, nil
}
