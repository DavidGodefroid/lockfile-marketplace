package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const goMod = `module github.com/DavidGodefroid/lockfile-marketplace/plugins/x.y

go 1.26.0

require github.com/pdfcpu/pdfcpu v0.16.2

replace github.com/DavidGodefroid/lockfile-marketplace/sdk => ../../sdk
`

const goSum = "github.com/pdfcpu/pdfcpu v0.16.2 h1:x=\ngithub.com/pdfcpu/pdfcpu v0.16.2/go.mod h1:y=\n"

// repo writes a repository with one plugin x.y whose main.go has the given
// imports, and returns its root.
func repo(t *testing.T, imports []string, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "plugins", "x.y")
	os.MkdirAll(dir, 0o755)
	var src strings.Builder
	src.WriteString("package main\n\nimport (\n")
	for _, i := range imports {
		src.WriteString("\t_ \"" + i + "\"\n")
	}
	src.WriteString(")\n\nfunc main() {}\n")
	all := map[string]string{"main.go": src.String(), "go.mod": goMod, "go.sum": goSum}
	for k, v := range files {
		all[k] = v
	}
	for name, content := range all {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func wantViolation(t *testing.T, root, substr string) {
	t.Helper()
	v, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range v {
		if strings.Contains(s, substr) {
			return
		}
	}
	t.Fatalf("no violation mentioning %q in %v", substr, v)
}

func TestImportCheckRejectsNet(t *testing.T) {
	for _, imp := range []string{"net", "net/http", "os/exec", "plugin", "unsafe", "syscall", "golang.org/x/sys/unix"} {
		wantViolation(t, repo(t, []string{imp}, nil), imp)
	}
}

func TestImportCheckAllowsSdk(t *testing.T) {
	root := repo(t, []string{
		"os", "io", "bytes",
		"github.com/DavidGodefroid/lockfile-marketplace/sdk",
		"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx",
		"github.com/pdfcpu/pdfcpu/pkg/api",
	}, nil)
	v, err := Check(root)
	if err != nil || len(v) != 0 {
		t.Fatalf("Check = %v, %v", v, err)
	}
}

func TestImportCheckIgnoresTests(t *testing.T) {
	root := repo(t, []string{"os"}, map[string]string{
		"plugin_test.go": "package main\n\nimport (\n\t_ \"os/exec\"\n\t_ \"net\"\n)\n",
	})
	if v, err := Check(root); err != nil || len(v) != 0 {
		t.Fatalf("Check = %v, %v", v, err)
	}
}

func TestImportCheckSubpackages(t *testing.T) {
	root := repo(t, []string{"os"}, map[string]string{})
	sub := filepath.Join(root, "plugins", "x.y", "internal", "z")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "z.go"), []byte("package z\n\nimport _ \"net\"\n"), 0o644)
	wantViolation(t, root, "net")
}

func TestImportCheckUnpinnedModule(t *testing.T) {
	root := repo(t, []string{"os"}, map[string]string{"go.sum": ""})
	wantViolation(t, root, "github.com/pdfcpu/pdfcpu")
}

func TestImportCheckForeignReplace(t *testing.T) {
	root := repo(t, []string{"os"}, map[string]string{
		"go.mod": goMod + "replace github.com/pdfcpu/pdfcpu => github.com/someone/fork v1.0.0\n",
	})
	wantViolation(t, root, "replace")
	root = repo(t, []string{"os"}, map[string]string{
		"go.mod": goMod + "replace github.com/pdfcpu/pdfcpu => ../../../elsewhere\n",
	})
	wantViolation(t, root, "replace")
}

func TestImportCheckUnparsable(t *testing.T) {
	root := repo(t, []string{"os"}, map[string]string{"broken.go": "package main\nimport (\n"})
	if _, err := Check(root); err == nil {
		t.Fatal("no error on a file that does not parse")
	}
}
