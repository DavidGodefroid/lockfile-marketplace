// Command pptx.has-hidden-slide answers: does the PowerPoint presentation
// contain a hidden slide?
package main

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

var (
	// errNoPresentation is returned for a package that is not a
	// PowerPoint presentation.
	errNoPresentation = errors.New("pptx.has-hidden-slide: no PresentationML presentation part")
	// errMissingSlide is returned when a slide relationship targets a part
	// absent from the package.
	errMissingSlide = errors.New("pptx.has-hidden-slide: slide part missing")
	// errNotSlide is returned for a slide part whose root is not p:sld.
	errNotSlide = errors.New("pptx.has-hidden-slide: slide part root is not p:sld")
)

// Content types of an XML presentation part: presentation, slide show,
// template and their macro-enabled variants.
var presentationTypes = []string{
	"application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml",
	"application/vnd.openxmlformats-officedocument.presentationml.slideshow.main+xml",
	"application/vnd.openxmlformats-officedocument.presentationml.template.main+xml",
	"application/vnd.ms-powerpoint.presentation.macroEnabled.main+xml",
	"application/vnd.ms-powerpoint.slideshow.macroEnabled.main+xml",
	"application/vnd.ms-powerpoint.template.macroEnabled.main+xml",
}

// Namespaces of PresentationML, transitional and strict.
const (
	presML       = "http://schemas.openxmlformats.org/presentationml/2006/main"
	presMLStrict = "http://purl.oclc.org/ooxml/presentationml/main"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		pres, err := presentation(p)
		if err != nil {
			return false, err
		}
		slides, err := slideParts(p, pres)
		if err != nil {
			return false, err
		}
		// Every slide is read to the end, so a malformed slide is an error
		// even after a hidden one.
		found := false
		for _, s := range slides {
			hidden, err := isHidden(s)
			if err != nil {
				return false, err
			}
			found = found || hidden
		}
		return found, nil
	})
}

// presentation returns the name of the part the root officeDocument
// relationship targets, when it exists and has a presentation content type.
func presentation(p *ooxml.Package) (string, error) {
	root := p.Part("_rels/.rels")
	if root == nil {
		return "", errNoPresentation
	}
	rels, err := ooxml.Rels(root)
	if err != nil {
		return "", err
	}
	for _, r := range rels {
		if !r.IsType("officeDocument") || strings.EqualFold(r.TargetMode, "External") {
			continue
		}
		name := ooxml.RootTarget(r.Target)
		if p.Part(name) == nil {
			continue
		}
		ct := p.ContentType(name)
		for _, t := range presentationTypes {
			if strings.EqualFold(ct, t) {
				return name, nil
			}
		}
	}
	return "", errNoPresentation
}

// slideParts returns the parts targeted by the slide relationships of the
// presentation part. A presentation without a rels part has no slides.
func slideParts(p *ooxml.Package, pres string) ([]*zip.File, error) {
	dir, base := path.Split(pres)
	relsPart := p.Part(dir + "_rels/" + base + ".rels")
	if relsPart == nil {
		return nil, nil
	}
	rels, err := ooxml.Rels(relsPart)
	if err != nil {
		return nil, err
	}
	var out []*zip.File
	for _, r := range rels {
		if !r.IsType("slide") || strings.EqualFold(r.TargetMode, "External") {
			continue
		}
		name := r.Target
		if !strings.HasPrefix(name, "/") {
			name = path.Join("/"+dir, name)
		}
		e := p.Part(ooxml.RootTarget(name))
		if e == nil {
			return nil, errMissingSlide
		}
		out = append(out, e)
	}
	return out, nil
}

// isHidden reads a whole slide part and reports whether its p:sld root has
// show set to false. The part must be well-formed throughout.
func isHidden(e *zip.File) (bool, error) {
	seenRoot, isSld, hidden := false, false, false
	err := ooxml.Walk(e, func(se xml.StartElement) bool {
		if seenRoot {
			return false
		}
		seenRoot = true
		isSld = (se.Name.Space == presML || se.Name.Space == presMLStrict) && se.Name.Local == "sld"
		if !isSld {
			return true
		}
		switch show(se) {
		case "0", "false":
			hidden = true
		}
		return false
	})
	if err != nil {
		return false, err
	}
	if !isSld {
		return false, errNotSlide
	}
	return hidden, nil
}

// show returns the unqualified show attribute of a slide root, with the
// surrounding white space an xsd:boolean allows removed.
func show(se xml.StartElement) string {
	for _, a := range se.Attr {
		if a.Name.Space == "" && a.Name.Local == "show" {
			return strings.TrimSpace(a.Value)
		}
	}
	return ""
}
