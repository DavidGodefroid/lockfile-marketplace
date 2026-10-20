// Command zip.has-path-traversal answers: does the ZIP archive have an entry
// whose name escapes the extraction directory?
package main

import (
	"os"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		r, err := zipx.Open(f)
		if err != nil {
			return false, err
		}
		for _, e := range r.File {
			if traverses(e.Name) {
				return true, nil
			}
		}
		return false, nil
	})
}

// traverses reports whether name is absolute (leading / or \), starts with
// a drive letter, or has a ".." component; both / and \ separate components,
// as extractors on Windows treat them alike.
func traverses(name string) bool {
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) {
		return true
	}
	if len(name) >= 2 && name[1] == ':' && isLetter(name[0]) {
		return true
	}
	for _, c := range strings.FieldsFunc(name, func(r rune) bool { return r == '/' || r == '\\' }) {
		if c == ".." {
			return true
		}
	}
	return false
}

func isLetter(b byte) bool {
	return 'a' <= b && b <= 'z' || 'A' <= b && b <= 'Z'
}
