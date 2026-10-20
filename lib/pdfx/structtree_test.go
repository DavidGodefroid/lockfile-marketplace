package pdfx

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// tagged returns a document whose structure tree root is {{obj 0}}, with
// root entries rootEntries, followed by the objects elems ({{obj 1}}…).
func tagged(rootEntries string, elems ...string) pdfxtest.Doc {
	return pdfxtest.Doc{
		Pages: []pdfxtest.Page{{}},
		Catalog: map[string]string{
			"/MarkInfo":       "<< /Marked true >>",
			"/StructTreeRoot": "{{obj 0}}",
		},
		Objects: append([]string{"<< /Type /StructTreeRoot" + rootEntries + " >>"}, elems...),
	}
}

func roles(ctx *model.Context) ([]string, error) {
	var got []string
	err := EachStructElem(ctx, func(_ types.Dict, role string) bool {
		got = append(got, role)
		return false
	})
	return got, err
}

func TestEachStructElemUntagged(t *testing.T) {
	got, err := roles(opened(t, pdfxtest.Doc{}))
	if err != nil || got != nil {
		t.Fatalf("got %v, %v", got, err)
	}
	got, err = roles(opened(t, tagged("")))
	if err != nil || got != nil {
		t.Fatalf("no /K: got %v, %v", got, err)
	}
}

func TestEachStructElemWalk(t *testing.T) {
	ctx := opened(t, tagged(" /K {{obj 1}}",
		"<< /Type /StructElem /S /Document /P {{obj 0}} /K [{{obj 2}} {{obj 3}} << /Type /MCR /Pg 3 0 R /MCID 0 >> 4] >>",
		"<< /Type /StructElem /S /Figure /P {{obj 1}} >>",
		"<< /S /P /P {{obj 1}} /K [<< /S /Span /P {{obj 3}} >> 1] >>",
	))
	got, err := roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Document", "Figure", "P", "Span"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEachStructElemRoleMap(t *testing.T) {
	ctx := opened(t, tagged(" /RoleMap << /Pic /Image /Image /Figure /Figure /Figure >> /K [{{obj 1}} {{obj 2}} {{obj 3}} {{obj 4}}]",
		"<< /S /Pic /P {{obj 0}} >>",
		"<< /S /Image /P {{obj 0}} >>",
		"<< /S /Figure /P {{obj 0}} >>",
		"<< /S /Note /P {{obj 0}} >>",
	))
	got, err := roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Figure", "Figure", "Figure", "Note"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEachStructElemRoleMapCycle(t *testing.T) {
	ctx := opened(t, tagged(" /RoleMap << /Pic /Image /Image /Pic >> /K {{obj 1}}", "<< /S /Pic /P {{obj 0}} >>"))
	if _, err := roles(ctx); !errors.Is(err, ErrRoleMapCycle) {
		t.Fatalf("err = %v", err)
	}
}

// An element listed under two parents is visited once.
func TestEachStructElemShared(t *testing.T) {
	ctx := readOnly(t, tagged(" /K [{{obj 1}} {{obj 2}}]",
		"<< /S /Sect /P {{obj 0}} /K {{obj 3}} >>",
		"<< /S /Div /P {{obj 0}} /K {{obj 3}} >>",
		"<< /S /Figure /P {{obj 1}} >>",
	))
	got, err := roles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Sect", "Figure", "Div"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEachStructElemCycle(t *testing.T) {
	ctx := readOnly(t, tagged(" /K {{obj 1}}",
		"<< /S /Sect /P {{obj 0}} /K {{obj 2}} >>",
		"<< /S /Div /P {{obj 1}} /K [{{obj 1}}] >>",
	))
	if _, err := roles(ctx); !errors.Is(err, ErrStructTreeCycle) {
		t.Fatalf("err = %v", err)
	}
}

// chain returns a tree of n nested elements.
func chain(n int) pdfxtest.Doc {
	elems := make([]string, n)
	for i := range elems {
		k := ""
		if i+1 < n {
			k = fmt.Sprintf(" /K {{obj %d}}", i+2)
		}
		elems[i] = fmt.Sprintf("<< /S /Div /P {{obj %d}}%s >>", i, k)
	}
	return tagged(" /K {{obj 1}}", elems...)
}

func TestEachStructElemDepth(t *testing.T) {
	got, err := roles(readOnly(t, chain(maxTreeDepth)))
	if err != nil || len(got) != maxTreeDepth {
		t.Fatalf("%d deep: %d visited, %v", maxTreeDepth, len(got), err)
	}
	if _, err := roles(readOnly(t, chain(maxTreeDepth+1))); !errors.Is(err, ErrStructTreeTooDeep) {
		t.Fatalf("%d deep: err = %v", maxTreeDepth+1, err)
	}
}

func TestEachStructElemStop(t *testing.T) {
	ctx := opened(t, tagged(" /K [{{obj 1}} {{obj 2}}]", "<< /S /Figure /P {{obj 0}} >>", "<< /S /P /P {{obj 0}} >>"))
	n := 0
	err := EachStructElem(ctx, func(_ types.Dict, role string) bool {
		n++
		return role == "Figure"
	})
	if err != nil || n != 1 {
		t.Fatalf("visited %d, %v", n, err)
	}
}

func TestEachStructElemBadS(t *testing.T) {
	ctx := readOnly(t, tagged(" /K {{obj 1}}", "<< /S (Figure) /P {{obj 0}} >>"))
	if _, err := roles(ctx); !errors.Is(err, ErrWrongType) {
		t.Fatalf("err = %v", err)
	}
	ctx = readOnly(t, tagged(" /RoleMap << /Pic (Figure) >> /K {{obj 1}}", "<< /S /Pic /P {{obj 0}} >>"))
	if _, err := roles(ctx); !errors.Is(err, ErrWrongType) {
		t.Fatalf("role map: err = %v", err)
	}
}

func TestEachStructElemObjStream(t *testing.T) {
	d := tagged(" /K {{obj 1}}", "<< /S /Figure /P {{obj 0}} >>")
	d.ObjStream = true
	got, err := roles(opened(t, d))
	if err != nil || strings.Join(got, ",") != "Figure" {
		t.Fatalf("got %v, %v", got, err)
	}
}
