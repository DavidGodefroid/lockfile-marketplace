// Command xlsx.has-formulas answers: does the Excel workbook contain
// formulas?
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

// errNoWorkbook is returned for a package that is not an Excel workbook.
var errNoWorkbook = errors.New("xlsx.has-formulas: no SpreadsheetML workbook part")

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
		sheets, err := worksheets(p, wb)
		if err != nil {
			return false, err
		}
		// Every worksheet is walked to its end, so that a malformed sheet
		// is an error whatever its place relative to a formula.
		found := false
		for _, e := range sheets {
			err := ooxml.Walk(e, func(se xml.StartElement) bool {
				if (se.Name.Space == sheetML || se.Name.Space == sheetMLStrict) && se.Name.Local == "f" {
					found = true
				}
				return false
			})
			if err != nil {
				return false, err
			}
		}
		return found, nil
	})
}

// workbook returns the name of the part the root officeDocument
// relationship targets, when it exists and has a workbook content type.
func workbook(p *ooxml.Package) (string, error) {
	root := p.Part("_rels/.rels")
	if root == nil {
		return "", errNoWorkbook
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
		for _, t := range workbookTypes {
			if strings.EqualFold(ct, t) {
				return name, nil
			}
		}
	}
	return "", errNoWorkbook
}

// worksheets returns the existing parts the workbook's worksheet
// relationships target. A workbook without a relationships part has none.
func worksheets(p *ooxml.Package, wb string) ([]*zip.File, error) {
	dir, file := path.Split(wb)
	rp := p.Part(dir + "_rels/" + file + ".rels")
	if rp == nil {
		return nil, nil
	}
	rels, err := ooxml.Rels(rp)
	if err != nil {
		return nil, err
	}
	var out []*zip.File
	for _, r := range rels {
		if !r.IsType("worksheet") || strings.EqualFold(r.TargetMode, "External") {
			continue
		}
		// A target is relative to the workbook's folder unless absolute.
		target := r.Target
		if !strings.HasPrefix(target, "/") {
			target = dir + target
		}
		if e := p.Part(ooxml.RootTarget(target)); e != nil {
			out = append(out, e)
		}
	}
	return out, nil
}
