package pdfx

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// readOnly reads without validation, for files validation would refuse.
func readOnly(t *testing.T, d pdfxtest.Doc) *model.Context {
	t.Helper()
	ctx, err := Read(bytes.NewReader(pdfxtest.Write(d)))
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestPageBoxInherited(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, MediaBox: &pdfxtest.Letter})
	r, err := PageBox(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if w, h := r.Size(); !near(w, 612) || !near(h, 792) {
		t.Fatalf("size %v x %v", w, h)
	}
}

func TestPageBoxCropClipped(t *testing.T) {
	crop := [4]float64{-100, 100, 300, 2000}
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: &crop}}})
	r, err := PageBox(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if r != (Rect{0, 100, 300, 841.89}) {
		t.Fatalf("box %+v", r)
	}
}

func TestPageBoxUserUnit(t *testing.T) {
	half := [4]float64{0, 0, 297.638, 420.945}
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{MediaBox: &half, UserUnit: 2}}})
	r, err := PageBox(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if w, h := r.Size(); !near(w, 595.276) || !near(h, 841.89) {
		t.Fatalf("size %v x %v", w, h)
	}
}

func TestPageBoxSwappedCoordinates(t *testing.T) {
	sw := [4]float64{595.276, 841.89, 0, 0}
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{MediaBox: &sw}}})
	r, err := PageBox(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if w, h := r.Size(); !near(w, 595.276) || !near(h, 841.89) {
		t.Fatalf("size %v x %v", w, h)
	}
}

func TestPageBoxMissingMediaBox(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, NoMediaBox: true})
	if _, err := PageBox(ctx, 1); !errors.Is(err, ErrMissingMediaBox) {
		t.Fatalf("err %v", err)
	}
}

func TestPageBoxEmptyCrop(t *testing.T) {
	crop := [4]float64{700, 900, 800, 1000}
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{CropBox: &crop}}})
	if _, err := PageBox(ctx, 1); !errors.Is(err, ErrEmptyCropBox) {
		t.Fatalf("err %v", err)
	}
}

func TestIsA4(t *testing.T) {
	for _, c := range []struct {
		w, h float64
		want bool
	}{
		{595.276, 841.890, true},
		{841.890, 595.276, true},
		{595.276 + 2.8, 841.890 - 2.8, true},
		{595.276 + 2.9, 841.890, false},
		{595.276, 841.890 + 2.9, false},
		{612, 792, false},
		{419.528, 595.276, false},
	} {
		if got := IsA4(c.w, c.h); got != c.want {
			t.Errorf("IsA4(%v, %v) = %v", c.w, c.h, got)
		}
	}
}
