// Command docx.has-comments answers: does the Word document contain
// comments?
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

// errNoDocument is returned for a package that is not a Word document.
var errNoDocument = errors.New("docx.has-comments: no word/document.xml")

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		if p.Part("word/document.xml") == nil {
			return false, errNoDocument
		}
		// Every part is read to the end, so a malformed part is an error
		// whatever the entry order.
		found := false
		for _, e := range p.Zip.File {
			// Part names are case-insensitive in OPC.
			if ok, _ := path.Match("word/*.xml", strings.ToLower(e.Name)); !ok {
				continue
			}
			err := ooxml.Walk(e, func(se xml.StartElement) bool {
				if ooxml.IsWordML(se.Name) && se.Name.Local == "comment" {
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
