package pdfxtest_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func TestWriteParses(t *testing.T) {
	rot := 90
	docs := map[string]pdfxtest.Doc{
		"one blank page": {Pages: []pdfxtest.Page{{}}},
		"text page": {Pages: []pdfxtest.Page{{
			Content:   "BT /F1 12 Tf 72 720 Td (sample) Tj ET",
			Resources: "<< /Font << /F1 {{obj 0}} >> >>",
		}}, Objects: []string{pdfxtest.Helvetica}},
		"page boxes": {Pages: []pdfxtest.Page{
			{MediaBox: &pdfxtest.Letter},
			{CropBox: &pdfxtest.A5, Rotate: &rot, UserUnit: 2},
		}},
		"pages rotate": {Pages: []pdfxtest.Page{{}, {}}, PagesRotate: &rot},
		"catalog and metadata": {
			Pages:    []pdfxtest.Page{{}},
			Catalog:  map[string]string{"/PageMode": "/UseNone"},
			Metadata: []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?><x:xmpmeta xmlns:x="adobe:ns:meta/"/><?xpacket end="w"?>`),
		},
		"annotation": {Pages: []pdfxtest.Page{{Annots: []string{"<< /Type /Annot /Subtype /Link /Rect [0 0 10 10] >>"}}}},
	}
	for name, d := range docs {
		t.Run(name, func(t *testing.T) {
			b := pdfxtest.Write(d)
			ctx, err := pdfx.Open(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("%v\n%s", err, b)
			}
			if ctx.PageCount != len(d.Pages) {
				t.Fatalf("PageCount %d", ctx.PageCount)
			}
		})
	}
}

func TestWriteParsesLayouts(t *testing.T) {
	base := pdfxtest.Pages("text", "image")
	base.Objects = append(base.Objects, "<< /S /JavaScript /JS (sample) >>")
	for name, mod := range map[string]func(*pdfxtest.Doc){
		"xref stream":               func(d *pdfxtest.Doc) { d.XRefStream = true },
		"object stream":             func(d *pdfxtest.Doc) { d.ObjStream = true },
		"hybrid":                    func(d *pdfxtest.Doc) { d.Hybrid = true },
		"linearized":                func(d *pdfxtest.Doc) { d.Linearized = true },
		"object stream, linearized": func(d *pdfxtest.Doc) { d.ObjStream, d.Linearized = true, true },
	} {
		t.Run(name, func(t *testing.T) {
			d := base
			mod(&d)
			b := pdfxtest.Write(d)
			ctx, err := pdfx.Open(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("%v\n%q", err, b)
			}
			if ctx.PageCount != 2 {
				t.Fatalf("PageCount %d", ctx.PageCount)
			}
			if got, err := pdfx.PagePaints(ctx, 1); err != nil || !got {
				t.Fatalf("page 1 paints = %v, %v", got, err)
			}
		})
	}
}

// lastStartXRef returns the offset after the last startxref keyword.
func lastStartXRef(t *testing.T, b []byte) int {
	t.Helper()
	i := bytes.LastIndex(b, []byte("startxref"))
	if i < 0 {
		t.Fatal("no startxref")
	}
	var off int
	if _, err := fmt.Sscan(string(b[i+len("startxref"):]), &off); err != nil {
		t.Fatal(err)
	}
	return off
}

// prevAt returns the /Prev of the section at off, or -1.
func prevAt(t *testing.T, b []byte, off int) int {
	t.Helper()
	end := bytes.Index(b[off:], []byte(">>"))
	sec := string(b[off : off+end])
	i := strings.Index(sec, "/Prev ")
	if i < 0 {
		return -1
	}
	var p int
	fmt.Sscan(sec[i+len("/Prev "):], &p)
	return p
}

func TestAppendUpdateChains(t *testing.T) {
	for _, xs := range []bool{false, true} {
		d := pdfxtest.Pages("text")
		d.XRefStream = xs
		orig := pdfxtest.Write(d)
		old := lastStartXRef(t, orig)
		upd := pdfxtest.AppendUpdate(orig, []string{"<< /Producer (sample) >>"}, "/Info {{obj 0}}")
		if !bytes.HasPrefix(upd, orig) {
			t.Fatal("update does not append")
		}
		newOff := lastStartXRef(t, upd)
		if newOff <= len(orig) || prevAt(t, upd, newOff) != old {
			t.Fatalf("xref stream %v: new section at %d, /Prev %d, want %d", xs, newOff, prevAt(t, upd, newOff), old)
		}
		if _, err := pdfx.Open(bytes.NewReader(upd)); err != nil {
			t.Fatalf("xref stream %v: %v", xs, err)
		}
		twice := pdfxtest.AppendUpdate(upd, nil, "")
		if prevAt(t, twice, lastStartXRef(t, twice)) != newOff {
			t.Fatal("second update does not chain to the first")
		}
	}
}

func TestLinearizedLayout(t *testing.T) {
	d := pdfxtest.Pages("text", "blank")
	d.Linearized = true
	b := pdfxtest.Write(d)
	first := bytes.Index(b, []byte(" obj"))
	if first < 0 || first > 1024 || !bytes.Contains(b[first:first+200], []byte("/Linearized")) {
		t.Fatalf("first object is not the linearization dict:\n%q", b[:200])
	}
	fp := lastStartXRef(t, b)
	main := prevAt(t, b, fp)
	if main <= fp || prevAt(t, b, main) != -1 {
		t.Fatalf("first-page section at %d, main at %d", fp, main)
	}
}

func TestHybrid(t *testing.T) {
	d := pdfxtest.Pages("text")
	d.Hybrid = true
	b := pdfxtest.Write(d)
	off := lastStartXRef(t, b)
	if !bytes.HasPrefix(b[off:], []byte("xref")) {
		t.Fatal("hybrid file must end with a classic table")
	}
	if !bytes.Contains(b[off:], []byte("/XRefStm ")) {
		t.Fatal("no /XRefStm in the trailer")
	}
}

func TestWriteLayout(t *testing.T) {
	b := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})
	if !bytes.HasPrefix(b, []byte("%PDF-1.7\n")) {
		t.Fatalf("header %q", b[:10])
	}
	if !bytes.HasSuffix(b, []byte("%%EOF\n")) {
		t.Fatalf("tail %q", b[len(b)-10:])
	}
	if !strings.Contains(string(b), "/MediaBox [0 0 595.276 841.89]") {
		t.Fatalf("no default A4 MediaBox on /Pages:\n%s", b)
	}
	if strings.Contains(string(pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, NoMediaBox: true})), "/MediaBox") {
		t.Fatal("NoMediaBox still writes a MediaBox")
	}
}

func TestStreamLength(t *testing.T) {
	s := pdfxtest.Stream("<< /Subtype /Image >>", []byte("abc"))
	if !strings.HasPrefix(s, "<< /Subtype /Image /Length 3 >>\nstream\nabc\nendstream") {
		t.Fatalf("stream %q", s)
	}
}

func TestPagesKinds(t *testing.T) {
	kinds := []string{"blank", "text", "white", "gray", "image", "invisible-text", "white-then-text"}
	paints := []bool{false, true, false, true, true, false, true}
	ctx, err := pdfx.Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Pages(kinds...))))
	if err != nil {
		t.Fatal(err)
	}
	if ctx.PageCount != len(kinds) {
		t.Fatalf("PageCount %d", ctx.PageCount)
	}
	for i, k := range kinds {
		got, err := pdfx.PagePaints(ctx, i+1)
		if err != nil || got != paints[i] {
			t.Errorf("%s: PagePaints = %v, %v", k, got, err)
		}
	}
}

func TestPagesUnknownKindPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("no panic")
		}
	}()
	pdfxtest.Pages("nonsense")
}

func TestEncrypt(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})
	for _, user := range []string{"", "sample-user"} {
		enc := pdfxtest.Encrypt(t, plain, user, "sample-owner")
		if bytes.Equal(enc, plain) || !bytes.Contains(enc, []byte("/Encrypt")) {
			t.Fatalf("user %q: not encrypted", user)
		}
	}
}

func TestVersionAndInfo(t *testing.T) {
	for _, layout := range []func(*pdfxtest.Doc){
		func(*pdfxtest.Doc) {},
		func(d *pdfxtest.Doc) { d.XRefStream = true },
		func(d *pdfxtest.Doc) { d.Linearized = true },
	} {
		d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Version: "2.0", Info: map[string]string{"/Title": "(sample)"}}
		layout(&d)
		b := pdfxtest.Write(d)
		if !bytes.HasPrefix(b, []byte("%PDF-2.0\n")) {
			t.Fatalf("header %q", b[:9])
		}
		ctx, err := pdfx.Open(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("%v\n%q", err, b)
		}
		if ctx.Info == nil {
			t.Fatal("no /Info in the trailer")
		}
	}
	if b := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}); !bytes.HasPrefix(b, []byte("%PDF-1.7\n")) || bytes.Contains(b, []byte("/Info")) {
		t.Fatal("default header or unexpected /Info")
	}
}
