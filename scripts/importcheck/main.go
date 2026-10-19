// Command importcheck fails when plugin code reaches for what a plugin must
// not have: the network, other programs, Go plugins, unsafe memory or raw
// system calls. It reads the non-test Go files under plugins/ (the sdk,
// which wraps the protocol, is not under plugins/ and is exempt), and it
// checks that each plugin module pins every requirement in its go.sum and
// replaces nothing but the in-repository sdk and lib modules.
//
//	go run ./scripts/importcheck [repository root]
package main

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

// denied lists import paths, and path prefixes ending in "/", that plugin
// code may not import.
var denied = []string{"net", "net/", "os/exec", "plugin", "unsafe", "syscall", "syscall/", "golang.org/x/sys/"}

// inRepo are the only modules a plugin may replace, with a relative path.
var inRepo = map[string]bool{
	"github.com/DavidGodefroid/lockfile-marketplace/sdk":      true,
	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx": true,
}

func isDenied(path string) bool {
	for _, d := range denied {
		if strings.HasSuffix(d, "/") {
			if strings.HasPrefix(path, d) {
				return true
			}
		} else if path == d {
			return true
		}
	}
	return false
}

// Check returns one line per violation under root/plugins. An error means
// the check itself could not run (a file that does not parse, an unreadable
// go.mod).
func Check(root string) ([]string, error) {
	var out []string
	plugins := filepath.Join(root, "plugins")
	err := filepath.WalkDir(plugins, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		switch {
		case d.Name() == "go.mod":
			v, err := checkModule(p)
			if err != nil {
				return err
			}
			out = append(out, v...)
		case strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go"):
			f, err := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imp := range f.Imports {
				path, _ := strconv.Unquote(imp.Path.Value)
				if isDenied(path) {
					out = append(out, fmt.Sprintf("%s: imports %s", rel, path))
				}
			}
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func checkModule(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	mf, err := modfile.Parse(path, b, nil)
	if err != nil {
		return nil, err
	}
	sums, err := readSums(filepath.Join(filepath.Dir(path), "go.sum"))
	if err != nil {
		return nil, err
	}
	var out []string
	replaced := map[string]bool{}
	for _, r := range mf.Replace {
		replaced[r.Old.Path] = true
		if !inRepo[r.Old.Path] || r.New.Version != "" || !strings.HasPrefix(r.New.Path, "../") ||
			strings.Count(r.New.Path, "../") > 2 {
			out = append(out, fmt.Sprintf("%s: replace %s => %s %s is not allowed", path, r.Old.Path, r.New.Path, r.New.Version))
		}
	}
	for _, r := range mf.Require {
		if replaced[r.Mod.Path] && inRepo[r.Mod.Path] {
			continue
		}
		if !sums[r.Mod.Path+" "+r.Mod.Version] {
			out = append(out, fmt.Sprintf("%s: %s %s is not pinned in go.sum", path, r.Mod.Path, r.Mod.Version))
		}
	}
	return out, nil
}

// readSums returns the "module version" pairs that have a content hash.
func readSums(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sums := map[string]bool{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 3 && !strings.HasSuffix(fields[1], "/go.mod") {
			sums[fields[0]+" "+fields[1]] = true
		}
	}
	return sums, s.Err()
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	v, err := Check(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "importcheck:", err)
		os.Exit(2)
	}
	for _, line := range v {
		fmt.Println(line)
	}
	if len(v) > 0 {
		os.Exit(1)
	}
}
