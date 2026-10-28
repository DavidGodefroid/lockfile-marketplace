package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk/harness"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestConformance(t *testing.T) {
	writeFixtures(t)
	harness.Check(t, ".")
}

func TestNoAnswer(t *testing.T) {
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":                nil,
		"sample.txt":           fx.Sample,
		"sample.docx":          fx.Docx(fx.Paragraph),
		"unclosed":             []byte("<!DOCTYPE r><r>"),
		"two-roots":            []byte("<!DOCTYPE a><a/><b/>"),
		"text-after-root":      []byte("<a/>tail"),
		"doctype-after-root":   []byte("<a/><!DOCTYPE a>"),
		"two-doctypes":         []byte("<!DOCTYPE a><!DOCTYPE a><a/>"),
		"lowercase-doctype":    []byte("<!doctype a><a/>"),
		"other-directive":      []byte("<!ELEMENT a ANY><a/>"),
		"doctype-no-name":      []byte("<!DOCTYPE><a/>"),
		"late-declaration":     []byte("<!DOCTYPE a><?xml version=\"1.0\"?><a/>"),
		"undefined-entity":     []byte("<!DOCTYPE a><a>&nbsp;</a>"),
		"invalid-utf8":         []byte("<!DOCTYPE a><a>\xff</a>"),
		"unsupported-charset":  []byte("<?xml version=\"1.0\" encoding=\"windows-1252\"?><!DOCTYPE a><a/>"),
		"unterminated-doctype": []byte("<!DOCTYPE a [<!ELEMENT a ANY>"),
		"doctype-without-root": []byte("<!DOCTYPE a>"),
		"double-bom":           []byte("\xef\xbb\xbf\xef\xbb\xbf<!DOCTYPE a><a/>"),
	})
}

type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }

func TestReadErrorIsReturned(t *testing.T) {
	want := errors.New("disk failure")
	ok, err := hasDoctype(io.MultiReader(strings.NewReader("<!DOCTYPE a><a>"), failingReader{want}))
	if !errors.Is(err, want) || ok {
		t.Fatalf("hasDoctype = %v, %v; want false, %v", ok, err, want)
	}
}

func TestMalformedIsNotXML(t *testing.T) {
	ok, err := hasDoctype(strings.NewReader("<a><b></a></b>"))
	if !errors.Is(err, errNotXML) || ok {
		t.Fatalf("hasDoctype = %v, %v; want false, %v", ok, err, errNotXML)
	}
}
