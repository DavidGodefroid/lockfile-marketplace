// Command xlsx.has-external-links answers: does the workbook link to
// external workbooks?
package main

import (
	"errors"
	"os"
	"path"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// errNotWorkbook is returned for a package whose main part is not a
// SpreadsheetML workbook.
var errNotWorkbook = errors.New("xlsx: no SpreadsheetML workbook part")

// workbookTypes are the content types of a workbook main part: workbook and
// template, plain and macro-enabled.
var workbookTypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml":    true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.template.main+xml": true,
	"application/vnd.ms-excel.sheet.macroEnabled.main+xml":                          true,
	"application/vnd.ms-excel.template.macroEnabled.main+xml":                       true,
}

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
		dir, base := path.Split(wb)
		e := p.Part(dir + "_rels/" + base + ".rels")
		if e == nil {
			return false, nil
		}
		rels, err := ooxml.Rels(e)
		if err != nil {
			return false, err
		}
		for _, r := range rels {
			if r.IsType("externalLink") {
				return true, nil
			}
		}
		return false, nil
	})
}

// workbook returns the part name of the workbook: the target of the
// package root's officeDocument relationship, which must exist and have a
// workbook content type.
func workbook(p *ooxml.Package) (string, error) {
	e := p.Part("_rels/.rels")
	if e == nil {
		return "", errNotWorkbook
	}
	rels, err := ooxml.Rels(e)
	if err != nil {
		return "", err
	}
	for _, r := range rels {
		if !r.IsType("officeDocument") {
			continue
		}
		t := ooxml.RootTarget(r.Target)
		if p.Part(t) != nil && workbookTypes[p.ContentType(t)] {
			return t, nil
		}
	}
	return "", errNotWorkbook
}
