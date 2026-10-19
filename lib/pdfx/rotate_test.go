package pdfx

import (
	"errors"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func rot(v int) *int { return &v }

func TestPageRotateInherited(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}, {Rotate: rot(180)}}, PagesRotate: rot(90)})
	for page, want := range map[int]int{1: 90, 2: 180} {
		got, err := PageRotate(ctx, page)
		if err != nil || got != want {
			t.Errorf("page %d: PageRotate = %d, %v; want %d", page, got, err, want)
		}
	}
}

func TestPageRotateAbsent(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})
	if got, err := PageRotate(ctx, 1); err != nil || got != 0 {
		t.Fatalf("PageRotate = %d, %v", got, err)
	}
}

func TestPageRotateRaw(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{Rotate: rot(-450)}}})
	if got, err := PageRotate(ctx, 1); err != nil || got != -450 {
		t.Fatalf("PageRotate = %d, %v", got, err)
	}
}

func TestNormRotate(t *testing.T) {
	for in, want := range map[int]int{0: 0, 90: 90, -90: 270, 360: 0, 450: 90, -450: 270, 180: 180, 270: 270} {
		if got, err := NormRotate(in); err != nil || got != want {
			t.Errorf("NormRotate(%d) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []int{45, -1, 91, 1000} {
		if _, err := NormRotate(in); !errors.Is(err, ErrInvalidRotate) {
			t.Errorf("NormRotate(%d) err %v", in, err)
		}
	}
}
