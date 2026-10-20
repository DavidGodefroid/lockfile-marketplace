// Command xlsx.has-hidden-sheet answers: does the Excel workbook contain a
// hidden sheet?
package main

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"os"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// errNoWorkbook is returned for a package that is not an Excel workbook.
var errNoWorkbook = errors.New("xlsx.has-hidden-sheet: no SpreadsheetML workbook part")

// Content types of an XML workbook part: workbook, template and their
// macro-enabled variants, and add-in.
var workbookTypes = []string{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.template.main+xml",
	"application/vnd.ms-excel.sheet.macroEnabled.main+xml",
	"application/vnd.ms-excel.template.macroEnabled.main+xml",
	"application/vnd.ms-excel.addin.macroEnabled.main+xml",
}

// Namespaces of SpreadsheetML, transitional and strict.
const (
	sheetML       = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	sheetMLStrict = "http://purl.oclc.org/ooxml/spreadsheetml/main"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		wb, err := workbook(p)
		if err != nil {
			return false, err
		}
		// The whole part is read, so a malformed workbook is an error even
		// after a hidden sheet.
		found := false
		err = ooxml.Walk(wb, func(se xml.StartElement) bool {
			if (se.Name.Space == sheetML || se.Name.Space == sheetMLStrict) && se.Name.Local == "sheet" {
				switch state(se) {
				case "hidden", "veryHidden":
					found = true
				}
			}
			return false
		})
		if err != nil {
			return false, err
		}
		return found, nil
	})
}

// workbook returns the part the root officeDocument relationship targets,
// when it exists and has a workbook content type.
func workbook(p *ooxml.Package) (*zip.File, error) {
	root := p.Part("_rels/.rels")
	if root == nil {
		return nil, errNoWorkbook
	}
	rels, err := ooxml.Rels(root)
	if err != nil {
		return nil, err
	}
	for _, r := range rels {
		if !r.IsType("officeDocument") || strings.EqualFold(r.TargetMode, "External") {
			continue
		}
		name := ooxml.RootTarget(r.Target)
		e := p.Part(name)
		if e == nil {
			continue
		}
		ct := p.ContentType(name)
		for _, t := range workbookTypes {
			if strings.EqualFold(ct, t) {
				return e, nil
			}
		}
	}
	return nil, errNoWorkbook
}

// state returns the unqualified state attribute of a sheet element.
func state(se xml.StartElement) string {
	for _, a := range se.Attr {
		if a.Name.Space == "" && a.Name.Local == "state" {
			return a.Value
		}
	}
	return ""
}
