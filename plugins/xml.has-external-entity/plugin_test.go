package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"testing/iotest"

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
		"empty":                 nil,
		"sample.txt":            fx.Sample,
		"sample.docx":           fx.Docx(fx.Paragraph),
		"unclosed":              []byte("<a><b></b>"),
		"two roots":             []byte("<a/><b/>"),
		"undeclared entity":     []byte("<a>&s;</a>"),
		"parameter entity ref":  []byte("<!DOCTYPE r [<!ENTITY % p \"x\">]><r>&p;</r>"),
		"directive after root":  []byte("<a/><!DOCTYPE a>"),
		"two doctypes":          []byte("<!DOCTYPE r><!DOCTYPE r><r/>"),
		"top-level entity":      []byte("<!ENTITY s SYSTEM \"sample.txt\"><r/>"),
		"unterminated subset":   []byte("<!DOCTYPE r [<!ENTITY s SYSTEM \"sample.txt\">><r/>"),
		"entity no literal":     []byte("<!DOCTYPE r [<!ENTITY s SYSTEM>]><r/>"),
		"entity bad keyword":    []byte("<!DOCTYPE r [<!ENTITY s FILE \"sample.txt\">]><r/>"),
		"parameter ndata":       []byte("<!DOCTYPE r [<!ENTITY % p SYSTEM \"sample\" NDATA n>]><r/>"),
		"garbage in subset":     []byte("<!DOCTYPE r [sample]><r/>"),
		"unsupported charset":   []byte("<?xml version=\"1.0\" encoding=\"windows-1252\"?><!DOCTYPE r [<!ENTITY s SYSTEM \"x\">]><r/>"),
		"external, then broken": []byte("<!DOCTYPE r [<!ENTITY s SYSTEM \"sample.txt\">]><r>"),
	})
}

func TestReadErrorIsReturned(t *testing.T) {
	want := errors.New("disk failure")
	r := io.MultiReader(strings.NewReader("<!DOCTYPE r [<!ENTITY s SYSTEM \"x\">]><r>"), iotest.ErrReader(want))
	ok, err := hasExternalEntity(r)
	if !errors.Is(err, want) || ok {
		t.Fatalf("hasExternalEntity = %v, %v; want false, %v", ok, err, want)
	}
}
