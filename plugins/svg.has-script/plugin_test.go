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
		"empty":                 nil,
		"sample.txt":            fx.Sample,
		"sample.docx":           fx.Docx(fx.Paragraph),
		"word xml":              fx.WordXML("document", fx.Paragraph),
		"html root":             []byte(`<html xmlns="http://www.w3.org/1999/xhtml"><script/></html>`),
		"svg without namespace": []byte(`<svg onload="f()"/>`),
		"svg wrong namespace":   []byte(`<svg xmlns="urn:sample"><script/></svg>`),
		"unclosed":              []byte(open + `<script>`),
		"script then malformed": svg(`<script/><a></b>`),
		"two roots":             []byte(open + `</svg>` + open + `</svg>`),
		"text after root":       []byte(open + `</svg>tail`),
		"undefined entity":      svg(`<text>&nbsp;</text>`),
		"invalid utf8":          svg("<text>\xff</text>"),
		"unsupported charset":   []byte(`<?xml version="1.0" encoding="windows-1252"?>` + open + `</svg>`),
	})
}

type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }

func TestReadErrorIsReturned(t *testing.T) {
	want := errors.New("disk failure")
	got, err := hasScript(io.MultiReader(strings.NewReader(open+"<script/>"), failingReader{want}))
	if !errors.Is(err, want) || got {
		t.Fatalf("hasScript = %v, %v; want false, %v", got, err, want)
	}
}
