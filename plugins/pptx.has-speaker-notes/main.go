// Command pptx.has-speaker-notes answers: does a slide of the PowerPoint
// presentation have speaker notes?
package main

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// Content types of a slide part and prefixes of a presentation main part
// (pptx, ppsx, potx, and their macro-enabled variants).
const (
	slideType = "application/vnd.openxmlformats-officedocument.presentationml.slide+xml"
	pmlPrefix = "application/vnd.openxmlformats-officedocument.presentationml."
	msPrefix  = "application/vnd.ms-powerpoint."
)

var errNotPresentation = errors.New("pptx.has-speaker-notes: not a PowerPoint presentation")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		if err := checkPresentation(p); err != nil {
			return false, err
		}
		for _, e := range p.Zip.File {
			if !strings.EqualFold(p.ContentType(e.Name), slideType) {
				continue
			}
			rels := p.Part(path.Join(path.Dir(e.Name), "_rels", path.Base(e.Name)+".rels"))
			if rels == nil {
				continue
			}
			rs, err := ooxml.Rels(rels)
			if err != nil {
				return false, err
			}
			for _, r := range rs {
				if r.IsType("notesSlide") {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

// checkPresentation returns an error unless the package's main part is a
// PresentationML presentation.
func checkPresentation(p *ooxml.Package) error {
	root := p.Part("_rels/.rels")
	if root == nil {
		return errNotPresentation
	}
	rs, err := ooxml.Rels(root)
	if err != nil {
		return err
	}
	for _, r := range rs {
		if !r.IsType("officeDocument") {
			continue
		}
		ct := strings.ToLower(p.ContentType(ooxml.RootTarget(r.Target)))
		if strings.HasPrefix(ct, pmlPrefix) || strings.HasPrefix(ct, msPrefix) {
			return nil
		}
	}
	return errNotPresentation
}
