// Package ooxml reads the structure of an Office Open XML package (DOCX,
// XLSX, PPTX and their macro-enabled variants): its parts, their content
// types and the relationships between them. It never reads a part's text.
package ooxml

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
)

// Namespaces of WordprocessingML, transitional and strict.
const (
	WordML       = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	WordMLStrict = "http://purl.oclc.org/ooxml/wordprocessingml/main"
)

// Bases of relationship types, transitional and strict.
var relBases = []string{
	"http://schemas.openxmlformats.org/officeDocument/2006/relationships/",
	"http://purl.oclc.org/ooxml/officeDocument/relationships/",
}

// ErrNotPackage is returned for a ZIP that has no [Content_Types].xml.
var ErrNotPackage = errors.New("ooxml: no [Content_Types].xml")

const contentTypesName = "[Content_Types].xml"

// Package is an opened OOXML package.
type Package struct {
	Zip       *zip.Reader
	defaults  map[string]string // extension (lower case) -> content type
	overrides map[string]string // part name (lower case, no leading /) -> content type
	parts     map[string]*zip.File
}

// Open opens f as a package: a ZIP whose [Content_Types].xml parses.
func Open(f *os.File) (*Package, error) {
	z, err := zipx.Open(f)
	if err != nil {
		return nil, err
	}
	p := &Package{Zip: z, defaults: map[string]string{}, overrides: map[string]string{}, parts: map[string]*zip.File{}}
	for _, e := range z.File {
		p.parts[strings.ToLower(e.Name)] = e
	}
	ct := p.Part(contentTypesName)
	if ct == nil {
		return nil, ErrNotPackage
	}
	err = decode(ct, func(se xml.StartElement) {
		switch se.Name.Local {
		case "Default":
			p.defaults[strings.ToLower(attr(se, "Extension"))] = attr(se, "ContentType")
		case "Override":
			p.overrides[partKey(attr(se, "PartName"))] = attr(se, "ContentType")
		}
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Part returns the entry of a part name (with or without a leading /,
// matched case-insensitively as OPC requires), or nil.
func (p *Package) Part(name string) *zip.File {
	return p.parts[partKey(name)]
}

// ContentType returns the content type of a part: its override, else the
// default for its extension, else "".
func (p *Package) ContentType(name string) string {
	if ct, ok := p.overrides[partKey(name)]; ok {
		return ct
	}
	return p.defaults[strings.ToLower(strings.TrimPrefix(path.Ext(name), "."))]
}

// ContentTypes returns every content type the package declares.
func (p *Package) ContentTypes() []string {
	var out []string
	for _, ct := range p.defaults {
		out = append(out, ct)
	}
	for _, ct := range p.overrides {
		out = append(out, ct)
	}
	return out
}

// Rel is one relationship of a .rels part.
type Rel struct {
	Type       string
	Target     string
	TargetMode string
}

// IsType reports whether the relationship type is the transitional or
// strict type with the given short name, such as "officeDocument".
func (r Rel) IsType(short string) bool {
	for _, b := range relBases {
		if r.Type == b+short {
			return true
		}
	}
	return false
}

// RelsParts returns the entries whose name ends in .rels.
func (p *Package) RelsParts() []*zip.File {
	var out []*zip.File
	for _, e := range p.Zip.File {
		if strings.HasSuffix(strings.ToLower(e.Name), ".rels") {
			out = append(out, e)
		}
	}
	return out
}

// Rels parses a .rels part.
func Rels(e *zip.File) ([]Rel, error) {
	var out []Rel
	err := decode(e, func(se xml.StartElement) {
		if se.Name.Local == "Relationship" {
			out = append(out, Rel{Type: attr(se, "Type"), Target: attr(se, "Target"), TargetMode: attr(se, "TargetMode")})
		}
	})
	return out, err
}

// RootTarget resolves the target of a relationship of the package root
// (_rels/.rels) to a part name.
func RootTarget(target string) string {
	return strings.TrimPrefix(path.Clean("/"+target), "/")
}

// IsWordML reports whether a name is in the WordprocessingML namespace.
func IsWordML(n xml.Name) bool {
	return n.Space == WordML || n.Space == WordMLStrict
}

// Walk streams the elements of an XML part to fn, never its text, with
// namespace prefixes resolved (Name.Space is the namespace URL). The part
// must be well-formed up to where fn returns true to stop.
func Walk(e *zip.File, fn func(xml.StartElement) bool) error {
	rc, err := e.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	d := xml.NewDecoder(rc)
	for {
		t, err := d.Token()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if se, ok := t.(xml.StartElement); ok && fn(se) {
			return nil
		}
	}
}

// decode reads a whole XML part, which must be well-formed, and hands
// every element to fn.
func decode(e *zip.File, fn func(xml.StartElement)) error {
	return Walk(e, func(se xml.StartElement) bool { fn(se); return false })
}

func attr(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local && a.Name.Space == "" {
			return a.Value
		}
	}
	return ""
}

func partKey(name string) string {
	return strings.ToLower(strings.TrimPrefix(name, "/"))
}
