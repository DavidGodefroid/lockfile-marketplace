// Command docx.has-document-protection answers: does the Word document
// declare document protection?
package main

import (
	"encoding/xml"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// errNotDocx is returned for a package whose main part is not a Word
// document.
var errNotDocx = errors.New("docx.has-document-protection: no WordprocessingML main document part")

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
		settings, err := settingsPart(p, doc)
		if err != nil || settings == "" {
			return false, err
		}
		// The settings part is read to its end, so a malformed part is an
		// error even after a documentProtection element.
		found := false
		err = ooxml.Walk(p.Part(settings), func(se xml.StartElement) bool {
			if ooxml.IsWordML(se.Name) && se.Name.Local == "documentProtection" {
				found = true
			}
			return false
		})
		if err != nil {
			return false, err
		}
		return found, nil
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

// settingsPart returns the name of the settings part of the main document
// doc, the internal target of its settings relationship that exists in the
// package, or "" when there is none.
func settingsPart(p *ooxml.Package, doc string) (string, error) {
	dir := path.Dir(doc)
	rels := p.Part(path.Join(dir, "_rels", path.Base(doc)+".rels"))
	if rels == nil {
		return "", nil
	}
	rs, err := ooxml.Rels(rels)
	if err != nil {
		return "", err
	}
	for _, r := range rs {
		if !r.IsType("settings") || strings.EqualFold(r.TargetMode, "External") {
			continue
		}
		// A target is relative to the source part's folder unless it
		// starts with a slash.
		name := ooxml.RootTarget(r.Target)
		if !strings.HasPrefix(r.Target, "/") {
			name = ooxml.RootTarget(path.Join(dir, r.Target))
		}
		if p.Part(name) != nil {
			return name, nil
		}
	}
	return "", nil
}
