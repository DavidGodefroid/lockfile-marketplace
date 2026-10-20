package pdfx

import (
	"bytes"
	"errors"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func opened(t *testing.T, d pdfxtest.Doc) *model.Context {
	t.Helper()
	if d.Pages == nil {
		d.Pages = []pdfxtest.Page{{}}
	}
	ctx, err := Open(bytes.NewReader(pdfxtest.Write(d)))
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func TestCatalogDict(t *testing.T) {
	ctx := opened(t, pdfxtest.Doc{
		Catalog: map[string]string{"/MarkInfo": "{{obj 0}}", "/Lang": "(en)"},
		Objects: []string{"<< /Marked true >>"},
	})
	d, err := CatalogDict(ctx, "MarkInfo")
	if err != nil || d == nil || d.BooleanEntry("Marked") == nil || !*d.BooleanEntry("Marked") {
		t.Fatalf("MarkInfo = %v, %v", d, err)
	}
	if d, err := CatalogDict(ctx, "StructTreeRoot"); d != nil || err != nil {
		t.Fatalf("absent = %v, %v", d, err)
	}
	if _, err := CatalogDict(ctx, "Lang"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("string as dict: %v", err)
	}
}

func TestCatalogArray(t *testing.T) {
	ctx := opened(t, pdfxtest.Doc{Catalog: map[string]string{
		"/OutputIntents": "[<< /Type /OutputIntent /S /GTS_PDFA1 >>]",
		"/PageLabels":    "<< /Nums [] >>",
	}})
	if a, err := CatalogArray(ctx, "OutputIntents"); err != nil || len(a) != 1 {
		t.Fatalf("OutputIntents = %v, %v", a, err)
	}
	if a, err := CatalogArray(ctx, "Threads"); a != nil || err != nil {
		t.Fatalf("absent = %v, %v", a, err)
	}
	if _, err := CatalogArray(ctx, "PageLabels"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("dict as array: %v", err)
	}
}

func TestNonEmptyStrings(t *testing.T) {
	for _, c := range []struct {
		raw  string
		want bool
	}{
		{"(sample)", true},
		{"()", false},
		{"(   )", false},
		{"<73616D706C65>", true},
		{"<>", false},
		{"<FEFF>", false},
		{"<FEFF0073>", true},
		{"<FEFF00200020>", false},
	} {
		ctx := opened(t, pdfxtest.Doc{
			Catalog: map[string]string{"/Lang": c.raw},
			Info:    map[string]string{"/Title": c.raw},
		})
		if got, err := CatalogNonEmptyString(ctx, "Lang"); err != nil || got != c.want {
			t.Errorf("catalog %s = %v, %v", c.raw, got, err)
		}
		if got, err := InfoHasNonEmpty(ctx, "Title"); err != nil || got != c.want {
			t.Errorf("info %s = %v, %v", c.raw, got, err)
		}
	}
}

func TestNonEmptyStringAbsentOrWrongType(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: map[string]string{"/Lang": "/en"}})
	if _, err := CatalogNonEmptyString(ctx, "Lang"); !errors.Is(err, ErrWrongType) {
		t.Fatalf("name as string: %v", err)
	}
	if got, err := InfoHasNonEmpty(ctx, "Title"); got || err != nil {
		t.Fatalf("no /Info = %v, %v", got, err)
	}
	if got, err := CatalogNonEmptyString(opened(t, pdfxtest.Doc{}), "Lang"); got || err != nil {
		t.Fatalf("absent = %v, %v", got, err)
	}
}

func TestInfoHas(t *testing.T) {
	ctx := opened(t, pdfxtest.Doc{Info: map[string]string{"/GTS_PDFXVersion": "(PDF/X-1:2001)"}})
	if got, err := InfoHas(ctx, "GTS_PDFXVersion"); !got || err != nil {
		t.Fatalf("present = %v, %v", got, err)
	}
	if got, err := InfoHas(ctx, "Title"); got || err != nil {
		t.Fatalf("absent = %v, %v", got, err)
	}
}

func TestHasCatalogMetadata(t *testing.T) {
	if got, err := HasCatalogMetadata(opened(t, pdfxtest.Doc{Metadata: pdfxtest.XMP(pdfxtest.NoPDFA)})); !got || err != nil {
		t.Fatalf("present = %v, %v", got, err)
	}
	if got, err := HasCatalogMetadata(opened(t, pdfxtest.Doc{})); got || err != nil {
		t.Fatalf("absent = %v, %v", got, err)
	}
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Catalog: map[string]string{"/Metadata": "<< /Type /Metadata >>"}})
	if _, err := HasCatalogMetadata(ctx); err == nil {
		t.Fatal("a dict is not a metadata stream")
	}
}
