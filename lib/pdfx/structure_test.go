package pdfx

import (
	"bytes"
	"errors"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func linearization(t *testing.T, b []byte) (bool, int64, error) {
	t.Helper()
	return Linearization(bytes.NewReader(b), int64(len(b)))
}

func TestLinearization(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})
	if ok, _, err := linearization(t, plain); ok || err != nil {
		t.Fatalf("plain = %v, %v", ok, err)
	}
	lin := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Linearized: true})
	ok, l, err := linearization(t, lin)
	if !ok || err != nil || l != int64(len(lin)) {
		t.Fatalf("linearized = %v, %d (size %d), %v", ok, l, len(lin), err)
	}
	upd := pdfxtest.AppendUpdate(lin, nil, "")
	if ok, l, err := linearization(t, upd); !ok || err != nil || l == int64(len(upd)) {
		t.Fatalf("updated = %v, %d, %v", ok, l, err)
	}
	bad := bytes.Replace(lin, []byte("/L "), []byte("/L -"), 1)
	if _, _, err := linearization(t, bad); !errors.Is(err, ErrBadLinearization) {
		t.Fatalf("negative /L: %v", err)
	}
}

func TestSameSize(t *testing.T) {
	a4 := Rect{0, 0, A4Short, A4Long}
	for _, c := range []struct {
		b    Rect
		want bool
	}{
		{Rect{0, 0, A4Long, A4Short}, true},
		{Rect{0, 0, A4Short + 1.4, A4Long - 1.4}, true},
		{Rect{0, 0, 612, 792}, false},
		{Rect{100, 100, 100 + A4Short, 100 + A4Long}, true},
	} {
		if got := SameSize(a4, c.b); got != c.want {
			t.Errorf("SameSize(A4, %v) = %v", c.b, got)
		}
	}
}

func TestOrientation(t *testing.T) {
	portrait := Rect{0, 0, A4Short, A4Long}
	landscape := Rect{0, 0, A4Long, A4Short}
	square := Rect{0, 0, 500, 501}
	for _, c := range []struct {
		box    Rect
		rotate int
		p, l   bool
	}{
		{portrait, 0, true, false},
		{portrait, 180, true, false},
		{portrait, 90, false, true},
		{landscape, 0, false, true},
		{landscape, 270, true, false},
		{square, 0, false, false},
		{square, 90, false, false},
	} {
		if p, l := Orientation(c.box, c.rotate); p != c.p || l != c.l {
			t.Errorf("Orientation(%v, %d) = %v, %v", c.box, c.rotate, p, l)
		}
	}
}
