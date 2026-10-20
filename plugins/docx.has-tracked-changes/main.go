// Command docx.has-tracked-changes answers: does the Word document contain
// tracked changes?
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

// errNoDocument is returned for a package with no word/document.xml.
var errNoDocument = errors.New("docx: no word/document.xml")

// revisions are the local names of WordprocessingML revision elements,
// besides the property changes whose names end in "Change".
var revisions = map[string]bool{
	"ins": true, "del": true, "moveFrom": true, "moveTo": true, "cellIns": true, "cellDel": true,
}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		if p.Part("word/document.xml") == nil {
			return false, errNoDocument
		}
		// Every part is walked to its end, so that a malformed part is an
		// error whatever its place in the ZIP relative to a revision.
		found := false
		for _, e := range p.Zip.File {
			if ok, _ := path.Match("word/*.xml", strings.ToLower(e.Name)); !ok {
				continue
			}
			err := ooxml.Walk(e, func(se xml.StartElement) bool {
				if ooxml.IsWordML(se.Name) && (revisions[se.Name.Local] || strings.HasSuffix(se.Name.Local, "Change")) {
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
