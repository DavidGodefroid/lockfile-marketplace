package pdfx

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func has(t *testing.T, b []byte) bool {
	t.Helper()
	ok, err := HasXMPProperty(b, PDFAIDNamespace, "part")
	if err != nil {
		t.Fatal(err)
	}
	return ok
}

func TestHasXMPPropertyElement(t *testing.T) {
	if !has(t, pdfxtest.XMP(pdfxtest.PDFAElement)) {
		t.Fatal("element not found")
	}
	if has(t, pdfxtest.XMP(pdfxtest.NoPDFA)) {
		t.Fatal("found in XMP without pdfaid")
	}
}

func TestHasXMPPropertyAttribute(t *testing.T) {
	if !has(t, pdfxtest.XMP(pdfxtest.PDFAAttribute)) {
		t.Fatal("attribute not found")
	}
}

func TestXMPNamespaceNotPrefix(t *testing.T) {
	if !has(t, pdfxtest.XMP(pdfxtest.PDFAOtherPrefix)) {
		t.Fatal("namespace under another prefix not found")
	}
	if has(t, pdfxtest.XMP(pdfxtest.PDFAWrongNS)) {
		t.Fatal("pdfaid prefix bound to another namespace counted")
	}
}

func TestXMPMalformed(t *testing.T) {
	if _, err := HasXMPProperty(pdfxtest.MalformedXMP, PDFAIDNamespace, "part"); !errors.Is(err, ErrXMPParse) {
		t.Fatalf("err %v", err)
	}
}

func TestXMPNoValueRetained(t *testing.T) {
	marker := "UNIQUE-SAMPLE-MARKER-7f3a"
	b := []byte(`<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">` +
		`<rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>` + marker + `</dc:title><broken attr=` + marker)
	_, err := HasXMPProperty(b, PDFAIDNamespace, "part")
	if err == nil || strings.Contains(err.Error(), marker) || err != ErrXMPParse {
		t.Fatalf("err %v", err)
	}
}

func TestToUTF8(t *testing.T) {
	x := pdfxtest.XMP(pdfxtest.PDFAElement)
	for name, in := range map[string][]byte{
		"utf-8":         x,
		"utf-8 bom":     append([]byte{0xEF, 0xBB, 0xBF}, x...),
		"utf-16 be bom": pdfxtest.UTF16BE(x),
		"utf-16 le bom": pdfxtest.UTF16LE(x),
	} {
		got, err := ToUTF8(in)
		if err != nil || !bytes.Equal(got, x) {
			t.Errorf("%s: %q, %v", name, got, err)
		}
	}
	if _, err := ToUTF8([]byte{0xFE, 0xFF, 0x00}); err == nil {
		t.Error("odd UTF-16 length accepted")
	}
}

func TestCatalogXMP(t *testing.T) {
	x := pdfxtest.XMP(pdfxtest.PDFAElement)
	ctx, err := Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Metadata: x})))
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := CatalogXMP(ctx, 1<<20)
	if err != nil || !ok || !bytes.Equal(got, x) {
		t.Fatalf("CatalogXMP = %q, %v, %v", got, ok, err)
	}
	if _, _, err := CatalogXMP(ctx, 10); !errors.Is(err, ErrXMPTooLarge) {
		t.Fatalf("cap: %v", err)
	}
	none, err := Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := CatalogXMP(none, 1<<20); ok || err != nil {
		t.Fatalf("no metadata: %v %v", ok, err)
	}
}
