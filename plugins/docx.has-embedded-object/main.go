// Command docx.has-embedded-object answers: does the Word document embed
// an object?
package main

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// errNotDocx is returned for a package whose main part is not a Word
// document.
var errNotDocx = errors.New("docx.has-embedded-object: no WordprocessingML main document part")

// mainTypes are the content types of a Word main document part: document,
// template and their macro-enabled variants.
var mainTypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.template.main+xml": true,
	"application/vnd.ms-word.document.macroEnabled.main+xml":                           true,
	"application/vnd.ms-word.template.macroEnabledTemplate.main+xml":                   true,
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		doc, err := mainPart(p)
		if err != nil {
			return false, err
		}
		rels := p.Part(path.Join(path.Dir(doc), "_rels", path.Base(doc)+".rels"))
		if rels == nil {
			return false, nil
		}
		rs, err := ooxml.Rels(rels)
		if err != nil {
			return false, err
		}
		for _, r := range rs {
			// A linked object points outside the package: it is not embedded.
			if (r.IsType("oleObject") || r.IsType("package")) && !strings.EqualFold(r.TargetMode, "External") {
				return true, nil
			}
		}
		return false, nil
	})
}

// mainPart returns the name of the Word main document part, the internal
// target of the root officeDocument relationship.
func mainPart(p *ooxml.Package) (string, error) {
	root := p.Part("_rels/.rels")
	if root == nil {
		return "", errNotDocx
	}
	rs, err := ooxml.Rels(root)
	if err != nil {
		return "", err
	}
	for _, r := range rs {
		if !r.IsType("officeDocument") || strings.EqualFold(r.TargetMode, "External") {
			continue
		}
		name := ooxml.RootTarget(r.Target)
		if p.Part(name) != nil && mainTypes[p.ContentType(name)] {
			return name, nil
		}
	}
	return "", errNotDocx
}
