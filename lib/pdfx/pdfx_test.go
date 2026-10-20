package pdfx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func open(t *testing.T, d pdfxtest.Doc) []byte {
	t.Helper()
	return pdfxtest.Write(d)
}

func TestOpenRejects(t *testing.T) {
	for name, b := range map[string][]byte{
		"empty":   nil,
		"text":    []byte("sample text, not a PDF\n"),
		"garbage": append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{0xff, 0x00, 'x'}, 300)...),
	} {
		if _, err := Open(bytes.NewReader(b)); err == nil {
			t.Errorf("%s: opened", name)
		}
	}
}

func page(content string) pdfxtest.Page { return pdfxtest.Page{Content: content} }

func TestPagePaints(t *testing.T) {
	cases := []struct {
		name    string
		content string
		res     string
		objs    []string
		paints  bool
	}{
		{name: "no content stream", paints: false},
		{name: "empty content", content: " ", paints: false},
		{name: "only state", content: "q 1 0 0 1 0 0 cm 0.5 g Q", paints: false},
		{name: "text", content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: []string{pdfxtest.Helvetica}, paints: true},
		{name: "TJ", content: "BT /F1 12 Tf [(sa) 10 (mple)] TJ ET", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: []string{pdfxtest.Helvetica}, paints: true},
		{name: "invisible text", content: "BT 3 Tr /F1 12 Tf (sample) Tj ET", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: []string{pdfxtest.Helvetica}, paints: false},
		{name: "full page white gray fill", content: "1 g 0 0 595.276 841.89 re f", paints: false},
		{name: "full page white rgb fill", content: "1 1 1 rg 0 0 600 850 re f*", paints: false},
		{name: "full page white cmyk fill", content: "0 0 0 0 k -10 -10 700 900 re F", paints: false},
		{name: "white fill not full page", content: "1 g 10 10 100 100 re f", paints: true},
		{name: "gray full page fill", content: "0.9 g 0 0 595.276 841.89 re f", paints: true},
		{name: "default black fill", content: "0 0 595.276 841.89 re f", paints: true},
		{name: "white fill under cm", content: "q 0.5 0 0 0.5 0 0 cm 1 g 0 0 595.276 841.89 re f Q", paints: true},
		{name: "white restored by Q", content: "1 g q 0 g Q 0 0 595.276 841.89 re f", paints: false},
		{name: "black restored by Q", content: "q 1 g Q 0 0 595.276 841.89 re f", paints: true},
		{name: "white non-rect path", content: "1 g 0 0 m 595 0 l 595 842 l h f", paints: true},
		{name: "stroke", content: "10 10 m 100 100 l S", paints: true},
		{name: "path discarded", content: "10 10 m 100 100 l n", paints: false},
		{name: "fill and stroke", content: "1 g 0 0 595.276 841.89 re B", paints: true},
		{name: "xobject", content: "q 100 0 0 100 0 0 cm /Im1 Do Q", res: "<< /XObject << /Im1 {{obj 0}} >> >>",
			objs: []string{pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0})}, paints: true},
		{name: "shading", content: "/Sh1 sh", paints: true},
		{name: "inline image", content: "BI /W 1 /H 1 /CS /G /BPC 8 ID \x00 EI", paints: true},
		{name: "operators inside strings are not operators", content: "BT 3 Tr (f S Tj) Tj ET", paints: false},
		{name: "white then text", content: "1 g 0 0 595.276 841.89 re f BT (sample) Tj ET", paints: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := page(c.content)
			p.Resources = c.res
			b := open(t, pdfxtest.Doc{Pages: []pdfxtest.Page{p}, Objects: c.objs})
			ctx, err := Open(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			got, err := PagePaints(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.paints {
				t.Fatalf("PagePaints = %v, want %v", got, c.paints)
			}
		})
	}
}

func TestPagePaintsBadPage(t *testing.T) {
	ctx, err := Open(bytes.NewReader(open(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PagePaints(ctx, 2); err == nil {
		t.Fatal("page 2 of 1")
	}
}

func TestPagePaintsCropBox(t *testing.T) {
	// White fill covering the crop box but not the media box: the visible
	// area is the crop box, so the page is blank.
	crop := [4]float64{100, 100, 300, 300}
	p := pdfxtest.Page{CropBox: &crop, Content: "1 g 100 100 200 200 re f"}
	ctx, err := Open(bytes.NewReader(open(t, pdfxtest.Doc{Pages: []pdfxtest.Page{p}})))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := PagePaints(ctx, 1); err != nil || got {
		t.Fatalf("PagePaints = %v, %v", got, err)
	}
}

func TestUsesUnembeddedFont(t *testing.T) {
	cases := []struct {
		name string
		res  string
		objs []string
		want bool
	}{
		{name: "no fonts", want: false},
		{name: "standard 14", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: []string{pdfxtest.Helvetica}, want: true},
		{name: "embedded truetype", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: pdfxtest.EmbeddedTrueType(0), want: false},
		{name: "truetype without font file", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: pdfxtest.UnembeddedTrueType(0), want: true},
		{name: "type0 embedded descendant", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: pdfxtest.Type0(0, true), want: false},
		{name: "type0 unembedded descendant", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: pdfxtest.Type0(0, false), want: true},
		{name: "type3", res: "<< /Font << /F1 {{obj 0}} >> >>", objs: []string{pdfxtest.Type3}, want: false},
		{name: "font inside a form xobject", res: "<< /XObject << /Fm1 {{obj 0}} >> >>",
			objs: []string{pdfxtest.Stream("<< /Type /XObject /Subtype /Form /BBox [0 0 10 10] /Resources << /Font << /F1 {{obj 1}} >> >> >>", []byte("BT (sample) Tj ET")), pdfxtest.Helvetica}, want: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := page("BT (sample) Tj ET")
			p.Resources = c.res
			ctx, err := Open(bytes.NewReader(open(t, pdfxtest.Doc{Pages: []pdfxtest.Page{p}, Objects: c.objs})))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			got, err := UsesUnembeddedFont(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("UsesUnembeddedFont = %v, want %v", got, c.want)
			}
		})
	}
}

func TestEncrypted(t *testing.T) {
	plain := open(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})
	if enc, err := Encrypted(bytes.NewReader(plain)); err != nil || enc {
		t.Fatalf("plain: %v %v", enc, err)
	}
	for _, user := range []string{"", "sample-user"} {
		b := pdfxtest.Encrypt(t, plain, user, "sample-owner")
		if enc, err := Encrypted(bytes.NewReader(b)); err != nil || !enc {
			t.Fatalf("user %q: %v %v", user, enc, err)
		}
	}
	if _, err := Encrypted(strings.NewReader("sample text")); err == nil {
		t.Fatal("text file has an answer")
	}
}

func TestHeader(t *testing.T) {
	for in, want := range map[string]bool{
		"%PDF-1.7\n":                         true,
		strings.Repeat("x", 1000) + "%PDF-1": true,
		strings.Repeat("x", 1020) + "%PDF-1": false,
		"":                                   false,
		"%PDF":                               false,
	} {
		got, err := HasHeader(strings.NewReader(in))
		if err != nil || got != want {
			t.Errorf("HasHeader(%d bytes) = %v, %v", len(in), got, err)
		}
	}
}

func TestLooksTruncated(t *testing.T) {
	good := open(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})
	cut := good[:len(good)-7]
	i := bytes.LastIndex(good, []byte("startxref\n"))
	badXref := append(bytes.Clone(good[:i]), []byte("startxref\n12\n%%EOF\n")...)
	outOfRange := append(bytes.Clone(good[:i]), []byte("startxref\n999999\n%%EOF\n")...)
	trailing := append(bytes.Clone(good), bytes.Repeat([]byte("\x00"), 2000)...)
	for name, c := range map[string]struct {
		b    []byte
		want bool
	}{
		"good":                 {good, false},
		"cut":                  {cut, true},
		"startxref to garbage": {badXref, true},
		"startxref past end":   {outOfRange, true},
		"2000 bytes after EOF": {trailing, true},
		"short trailing bytes": {append(bytes.Clone(good), "\r\n  "...), false},
	} {
		got, err := LooksTruncated(bytes.NewReader(c.b), int64(len(c.b)))
		if err != nil || got != c.want {
			t.Errorf("%s: LooksTruncated = %v, %v; want %v", name, got, err, c.want)
		}
	}
	if _, err := LooksTruncated(strings.NewReader("sample text"), 11); err == nil {
		t.Error("text file has an answer")
	}
}

func TestRead(t *testing.T) {
	// A PostScript XObject drawn from a page fails pdfcpu's validation.
	ps := pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{Content: "q /X0 Do Q", Resources: "<< /XObject << /X0 {{obj 0}} >> >>"}},
		Objects: []string{pdfxtest.Stream("<< /Type /XObject /Subtype /PS >>", []byte("sample"))},
	}
	b := pdfxtest.Write(ps)
	if _, err := Open(bytes.NewReader(b)); err == nil {
		t.Fatal("Open: want a validation error")
	}
	ctx, err := Read(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if ctx.PageCount != 1 {
		t.Errorf("PageCount = %d, want 1", ctx.PageCount)
	}
	for name, b := range map[string][]byte{
		"empty": nil,
		"text":  []byte("sample text, not a PDF\n"),
	} {
		if _, err := Read(bytes.NewReader(b)); err == nil {
			t.Errorf("%s: read", name)
		}
	}
}
