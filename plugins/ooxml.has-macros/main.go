// Command ooxml.has-macros answers: does the Office Open XML package
// contain macros?
package main

import (
	"os"
	"path"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// vbaProject is the content type of a VBA project part.
const vbaProject = "application/vnd.ms-office.vbaProject"

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		p, err := ooxml.Open(f)
		if err != nil {
			return false, err
		}
		// Only parts present in the ZIP count: a content type declared
		// for no existing part is not a macro.
		for _, e := range p.Zip.File {
			if strings.HasSuffix(e.Name, "/") {
				continue
			}
			if strings.EqualFold(path.Base(e.Name), "vbaProject.bin") {
				return true, nil
			}
			ct := p.ContentType(e.Name)
			if strings.EqualFold(ct, vbaProject) || strings.Contains(strings.ToLower(ct), "macrosheet") {
				return true, nil
			}
		}
		return false, nil
	})
}
