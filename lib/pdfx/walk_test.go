package pdfx

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func hasKey(key string) func(types.Object) bool {
	return func(o types.Object) bool {
		d, ok := o.(types.Dict)
		if !ok {
			return false
		}
		_, found := d.Find(key)
		return found
	}
}

func found(t *testing.T, d pdfxtest.Doc, key string) bool {
	t.Helper()
	ctx, err := Open(bytes.NewReader(pdfxtest.Write(d)))
	if err != nil {
		t.Fatal(err)
	}
	hit := false
	if err := EachObject(ctx, func(o types.Object) bool {
		if hasKey(key)(o) {
			hit = true
		}
		return hit
	}); err != nil {
		t.Fatal(err)
	}
	return hit
}

func TestEachObjectVisitsObjStream(t *testing.T) {
	d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{"<< /Marker /Here >>"}, ObjStream: true}
	if !found(t, d, "Marker") {
		t.Fatal("object inside the object stream not visited")
	}
}

func TestEachObjectNested(t *testing.T) {
	d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{"<< /A [1 << /B << /Marker 1 >> >>] >>"}}
	if !found(t, d, "Marker") {
		t.Fatal("nested inline dict not visited")
	}
	if found(t, pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}, "Marker") {
		t.Fatal("found a marker that is not there")
	}
}

func TestEachObjectStreamDict(t *testing.T) {
	d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{pdfxtest.Stream("<< /Marker 1 >>", []byte("x"))}}
	if !found(t, d, "Marker") {
		t.Fatal("stream dictionary not visited")
	}
}

func TestEachObjectCycle(t *testing.T) {
	d := pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{"<< /Self {{obj 0}} >>", "<< /A {{obj 2}} >>", "<< /B {{obj 1}} >>"}}
	if found(t, d, "Marker") {
		t.Fatal("found a marker that is not there")
	}
}

func TestEachObjectDepthCap(t *testing.T) {
	deep := strings.Repeat("[", MaxDepth+1) + strings.Repeat("]", MaxDepth+1)
	ctx, err := Read(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{deep}})))
	if err != nil {
		t.Skipf("pdfcpu refuses the nesting itself: %v", err)
	}
	if err := EachObject(ctx, func(types.Object) bool { return false }); !errors.Is(err, ErrTooDeep) {
		t.Fatalf("err %v", err)
	}
}

func TestEachObjectStop(t *testing.T) {
	ctx, err := Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Pages("text", "text"))))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	if err := EachObject(ctx, func(types.Object) bool { n++; return true }); err != nil || n != 1 {
		t.Fatalf("visited %d, err %v", n, err)
	}
	n = 0
	EachObject(ctx, func(types.Object) bool { n++; return false })
	if n < 10 {
		t.Fatalf("full walk visited only %d nodes", n)
	}
}
