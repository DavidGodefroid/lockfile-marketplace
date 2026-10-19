// Package pdfxtest writes synthetic PDF files for plugin tests, byte by
// byte, so a fixture can carry exactly the structure a test needs (page
// boxes, content operators, fonts, metadata, broken tails). Content is
// neutral ("sample"); no real document is ever used.
package pdfxtest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Page sizes in points.
var (
	A4     = [4]float64{0, 0, 595.276, 841.890}
	Letter = [4]float64{0, 0, 612, 792}
	A5     = [4]float64{0, 0, 419.528, 595.276}
)

// Page is one page. Raw strings are PDF syntax and may reference
// Doc.Objects with {{obj i}}.
type Page struct {
	MediaBox, CropBox *[4]float64
	UserUnit          float64
	Rotate            *int
	Annots            []string // raw annotation dicts
	Content           string   // raw content stream; empty = no /Contents
	Resources         string   // raw resources dict; empty = << >>
}

// Doc is one PDF file.
type Doc struct {
	Pages       []Page
	MediaBox    *[4]float64 // on /Pages, inherited; nil = A4
	NoMediaBox  bool        // no MediaBox on /Pages
	PagesRotate *int
	Catalog     map[string]string // extra raw catalog entries, key with its slash
	Objects     []string          // extra objects, raw; referenced as {{obj i}}
	Metadata    []byte            // catalog XMP metadata stream
}

// Raw fonts and font builders.
const (
	Helvetica = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"
	Type3     = "<< /Type /Font /Subtype /Type3 /FontBBox [0 0 1 1] /FontMatrix [0.001 0 0 0.001 0 0] " +
		"/CharProcs << >> /Encoding << /Type /Encoding /Differences [0] >> /FirstChar 0 /LastChar 0 /Widths [0] " +
		"/Resources << >> >>"
)

func descriptor(file string) string {
	d := "<< /Type /FontDescriptor /FontName /SampleSans /Flags 32 /FontBBox [0 0 1000 1000] " +
		"/ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 700 /StemV 80"
	if file != "" {
		d += " /FontFile2 " + file
	}
	return d + " >>"
}

func fontFile() string {
	return Stream("<< /Length1 19 >>", []byte("sample font program"))
}

// EmbeddedTrueType is a TrueType font with a FontFile2, as objects to put
// at index at of Doc.Objects (the font dict is Objects[at]).
func EmbeddedTrueType(at int) []string {
	return []string{
		fmt.Sprintf("<< /Type /Font /Subtype /TrueType /BaseFont /SampleSans /FirstChar 32 /LastChar 32 /Widths [250] /FontDescriptor {{obj %d}} >>", at+1),
		descriptor(fmt.Sprintf("{{obj %d}}", at+2)),
		fontFile(),
	}
}

// UnembeddedTrueType is a TrueType font whose descriptor has no font file.
func UnembeddedTrueType(at int) []string {
	return []string{
		fmt.Sprintf("<< /Type /Font /Subtype /TrueType /BaseFont /SampleSans /FirstChar 32 /LastChar 32 /Widths [250] /FontDescriptor {{obj %d}} >>", at+1),
		descriptor(""),
	}
}

// Type0 is a composite font with one CIDFontType2 descendant, embedded or
// not.
func Type0(at int, embedded bool) []string {
	objs := []string{
		fmt.Sprintf("<< /Type /Font /Subtype /Type0 /BaseFont /SampleCID /Encoding /Identity-H /DescendantFonts [{{obj %d}}] >>", at+1),
		fmt.Sprintf("<< /Type /Font /Subtype /CIDFontType2 /BaseFont /SampleCID /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor {{obj %d}} /CIDToGIDMap /Identity >>", at+2),
	}
	if embedded {
		return append(objs, descriptor(fmt.Sprintf("{{obj %d}}", at+3)), fontFile())
	}
	return append(objs, descriptor(""))
}

// Stream is a raw stream object: dict (which must not hold /Length) with
// /Length added, then data.
func Stream(dict string, data []byte) string {
	d := strings.TrimSpace(dict)
	d = strings.TrimSuffix(d, ">>")
	return fmt.Sprintf("%s /Length %d >>\nstream\n%s\nendstream", strings.TrimRight(d, " "), len(data), data)
}

func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func box(b [4]float64) string {
	return fmt.Sprintf("[%s %s %s %s]", num(b[0]), num(b[1]), num(b[2]), num(b[3]))
}

var objRef = regexp.MustCompile(`\{\{obj (\d+)\}\}`)

// Write builds the file: 1 catalog, 2 page tree, then each page and its
// content stream, then the metadata stream, then Doc.Objects; a classic
// xref table and trailer.
func Write(d Doc) []byte {
	var bodies []string
	add := func(s string) int { bodies = append(bodies, s); return len(bodies) }
	add("") // catalog, filled below
	add("") // pages
	var kids []string
	for _, p := range d.Pages {
		pageNo := add("")
		kids = append(kids, fmt.Sprintf("%d 0 R", pageNo))
		var b strings.Builder
		b.WriteString("<< /Type /Page /Parent 2 0 R")
		if p.MediaBox != nil {
			b.WriteString(" /MediaBox " + box(*p.MediaBox))
		}
		if p.CropBox != nil {
			b.WriteString(" /CropBox " + box(*p.CropBox))
		}
		if p.UserUnit != 0 {
			b.WriteString(" /UserUnit " + num(p.UserUnit))
		}
		if p.Rotate != nil {
			b.WriteString(" /Rotate " + strconv.Itoa(*p.Rotate))
		}
		res := p.Resources
		if res == "" {
			res = "<< >>"
		}
		b.WriteString(" /Resources " + res)
		if len(p.Annots) > 0 {
			b.WriteString(" /Annots [" + strings.Join(p.Annots, " ") + "]")
		}
		if p.Content != "" {
			c := add(Stream("<< >>", []byte(p.Content)))
			b.WriteString(fmt.Sprintf(" /Contents %d 0 R", c))
		}
		b.WriteString(" >>")
		bodies[pageNo-1] = b.String()
	}
	meta := 0
	if d.Metadata != nil {
		meta = add(Stream("<< /Type /Metadata /Subtype /XML >>", d.Metadata))
	}
	first := len(bodies) + 1
	for _, o := range d.Objects {
		add(o)
	}

	var cat strings.Builder
	cat.WriteString("<< /Type /Catalog /Pages 2 0 R")
	if meta != 0 {
		cat.WriteString(fmt.Sprintf(" /Metadata %d 0 R", meta))
	}
	for k, v := range d.Catalog {
		cat.WriteString(" " + k + " " + v)
	}
	cat.WriteString(" >>")
	bodies[0] = cat.String()

	var pages strings.Builder
	pages.WriteString(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d", strings.Join(kids, " "), len(d.Pages)))
	if !d.NoMediaBox {
		mb := A4
		if d.MediaBox != nil {
			mb = *d.MediaBox
		}
		pages.WriteString(" /MediaBox " + box(mb))
	}
	if d.PagesRotate != nil {
		pages.WriteString(" /Rotate " + strconv.Itoa(*d.PagesRotate))
	}
	pages.WriteString(" >>")
	bodies[1] = pages.String()

	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(bodies))
	for i, body := range bodies {
		body = objRef.ReplaceAllStringFunc(body, func(m string) string {
			n, _ := strconv.Atoi(objRef.FindStringSubmatch(m)[1])
			return fmt.Sprintf("%d 0 R", first+n)
		})
		offsets[i] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, body)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(bodies)+1)
	for _, off := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(bodies)+1, xref)
	return out.Bytes()
}

// Sample is a file that is not a PDF.
var Sample = []byte("sample text, not a PDF\n")

const (
	textOp = "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	fonts  = "<< /Font << /F1 {{obj 0}} >> /XObject << /Im1 {{obj 1}} >> >>"
)

// pageKinds are the page contents Pages knows.
var pageKinds = map[string]string{
	"blank":           "",
	"text":            textOp,
	"white":           "1 g 0 0 595.276 841.89 re f",
	"gray":            "0.9 g 0 0 595.276 841.89 re f",
	"image":           "q 100 0 0 100 72 600 cm /Im1 Do Q",
	"invisible-text":  "BT 3 Tr /F1 12 Tf 72 720 Td (sample) Tj ET",
	"white-then-text": "1 g 0 0 595.276 841.89 re f 0 g " + textOp,
}

// Pages is an A4 document with one page per kind: blank, text, white (a
// full-page white fill), gray (a full-page light gray fill), image,
// invisible-text, white-then-text. It panics on an unknown kind.
func Pages(kinds ...string) Doc {
	d := Doc{Objects: []string{
		Helvetica,
		Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0x80}),
	}}
	for _, k := range kinds {
		c, ok := pageKinds[k]
		if !ok {
			panic("pdfxtest: unknown page kind " + k)
		}
		d.Pages = append(d.Pages, Page{Content: c, Resources: fonts})
	}
	return d
}

// Encrypt encrypts pdf with pdfcpu (AES-256). An empty user password gives
// a file anyone can open.
func Encrypt(t *testing.T, pdf []byte, user, owner string) []byte {
	t.Helper()
	conf := model.NewStatelessConfiguration()
	conf.UserPW, conf.OwnerPW = user, owner
	conf.EncryptUsingAES, conf.EncryptKeyLength = true, 256
	var out bytes.Buffer
	if err := api.Encrypt(context.Background(), bytes.NewReader(pdf), &out, conf); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return out.Bytes()
}

// Raw writes b as dir/name and returns the path.
func Raw(t *testing.T, dir, name string, b []byte) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
