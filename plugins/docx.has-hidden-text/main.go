// Command docx.has-hidden-text answers: does the Word document contain
// hidden text?
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

// off are the ST_OnOff values of a w:val attribute that turn a toggle
// property off.
var off = map[string]bool{"0": true, "false": true, "off": true}

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		e := p.Part("word/document.xml")
		if e == nil {
			return false, errNoDocument
		}
		// The part is walked to its end, so that a malformed part is an
		// error whatever its place relative to a vanish element.
		found := false
		err = ooxml.Walk(e, func(se xml.StartElement) bool {
			if ooxml.IsWordML(se.Name) && se.Name.Local == "vanish" && !turnedOff(se) {
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

// turnedOff reports whether a toggle element has a w:val that sets it off.
func turnedOff(se xml.StartElement) bool {
	for _, a := range se.Attr {
		if ooxml.IsWordML(a.Name) && a.Name.Local == "val" {
			return off[a.Value]
		}
	}
	return false
}
