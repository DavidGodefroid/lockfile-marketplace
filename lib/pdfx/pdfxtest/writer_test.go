package pdfxtest_test

import (
	"bytes"
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
