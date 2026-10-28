// Command docx.has-field-codes answers: does the Word document contain
// field codes?
package main

import (
	"encoding/xml"
	"errors"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// errNoDocument is returned for a package with no word/document.xml.
var errNoDocument = errors.New("docx: no word/document.xml")

// fields are the local names of the WordprocessingML elements that carry a
// field instruction: a simple field, or the instruction run of a complex one.
var fields = map[string]bool{"fldSimple": true, "instrText": true}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		doc := p.Part("word/document.xml")
		if doc == nil {
			return false, errNoDocument
		}
		// The part is walked to its end, so that a malformed part is an
		// error even after a field.
		found := false
		err = ooxml.Walk(doc, func(se xml.StartElement) bool {
			if ooxml.IsWordML(se.Name) && fields[se.Name.Local] {
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
