package ooxml_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/ooxml"
)

func open(t *testing.T, b []byte) (*ooxml.Package, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return ooxml.Open(f)
}

func TestOpenDocx(t *testing.T) {
	p, err := open(t, formatxtest.Docx(formatxtest.Paragraph))
	if err != nil {
		t.Fatal(err)
	}
	if got := p.ContentType("/word/document.xml"); got != formatxtest.DocxMain {
		t.Errorf("override: got %q", got)
	}
	if got := p.ContentType("_rels/.rels"); got != "application/vnd.openxmlformats-package.relationships+xml" {
		t.Errorf("default: got %q", got)
	}
	if p.Part("/WORD/document.xml") == nil {
		t.Error("part lookup is not case-insensitive")
	}
	rels, err := ooxml.Rels(p.Part("_rels/.rels"))
	if err != nil || len(rels) != 1 || !rels[0].IsType("officeDocument") || rels[0].IsType("hyperlink") {
		t.Fatalf("rels: %+v, %v", rels, err)
	}
	if got := ooxml.RootTarget("/word/../word/document.xml"); got != "word/document.xml" {
		t.Errorf("RootTarget: %q", got)
	}
	if n := len(p.RelsParts()); n != 1 {
		t.Errorf("RelsParts: %d", n)
	}
}

func TestOpenRefuses(t *testing.T) {
	if _, err := open(t, formatxtest.Sample); err == nil {
		t.Error("not a zip: no error")
	}
	if _, err := open(t, formatxtest.Zip(formatxtest.File("a.txt", formatxtest.Sample))); !errors.Is(err, ooxml.ErrNotPackage) {
		t.Errorf("zip without content types: %v", err)
	}
	bad := formatxtest.Zip(formatxtest.File("[Content_Types].xml", []byte("<Types><Default></Types>")))
	if _, err := open(t, bad); err == nil {
		t.Error("malformed content types: no error")
	}
}
