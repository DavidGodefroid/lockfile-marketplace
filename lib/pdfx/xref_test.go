package pdfx

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func chainOf(t *testing.T, b []byte) []Section {
	t.Helper()
	c, err := XRefChain(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func linearized(t *testing.T, b []byte) bool {
	t.Helper()
	l, err := IsLinearized(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func doc(mod func(*pdfxtest.Doc)) []byte {
	d := pdfxtest.Pages("text", "blank")
	if mod != nil {
		mod(&d)
	}
	return pdfxtest.Write(d)
}

func update(b []byte) []byte {
	return pdfxtest.AppendUpdate(b, []string{"<< /Producer (sample) >>"}, "/Info {{obj 0}}")
}

func TestXRefChainSingle(t *testing.T) {
	c := chainOf(t, doc(nil))
	if len(c) != 1 || c[0].Stream || c[0].HasPrev || c[0].HasXRefStm {
		t.Fatalf("chain %+v", c)
	}
}

func TestXRefChainOneUpdate(t *testing.T) {
	orig := doc(nil)
	c := chainOf(t, update(orig))
	if len(c) != 2 || !c[0].HasPrev || c[0].Offset <= int64(len(orig))-1 || c[1].HasPrev || c[0].Prev != c[1].Offset {
		t.Fatalf("chain %+v", c)
	}
	if CountRevisions(c, false) != 2 {
		t.Fatal("revisions")
	}
}

func TestXRefChainTwoUpdates(t *testing.T) {
	if c := chainOf(t, update(update(doc(nil)))); len(c) != 3 {
		t.Fatalf("chain %+v", c)
	}
}

func TestXRefChainStream(t *testing.T) {
	c := chainOf(t, doc(func(d *pdfxtest.Doc) { d.XRefStream = true }))
	if len(c) != 1 || !c[0].Stream {
		t.Fatalf("chain %+v", c)
	}
}

func TestXRefChainStreamUpdate(t *testing.T) {
	c := chainOf(t, update(doc(func(d *pdfxtest.Doc) { d.XRefStream = true })))
	if len(c) != 2 || !c[0].Stream || !c[1].Stream {
		t.Fatalf("chain %+v", c)
	}
}

func TestXRefChainHybrid(t *testing.T) {
	b := doc(func(d *pdfxtest.Doc) { d.Hybrid = true })
	c := chainOf(t, b)
	if len(c) != 1 || !c[0].HasXRefStm || CountRevisions(c, linearized(t, b)) != 1 {
		t.Fatalf("chain %+v", c)
	}
}

// withPrev adds /Prev to the trailer of the original section of b, with a
// fixed-width placeholder, and returns the result and the placeholder.
func withPrev(b []byte) []byte {
	return bytes.Replace(b, []byte("/Root 1 0 R >>\nstartxref"), []byte("/Root 1 0 R /Prev @@PREVX@@@ >>\nstartxref"), 1)
}

func setPrev(b []byte, v int) []byte {
	s := fmt.Sprintf("%-10d", v)
	return bytes.Replace(b, []byte("@@PREVX@@@"), []byte(s), 1)
}

func TestXRefChainLoop(t *testing.T) {
	// Section pointing to itself.
	orig := doc(nil)
	self := withPrev(orig)
	i := bytes.LastIndex(orig, []byte("xref\n0 "))
	self = setPrev(self, i)
	if _, err := XRefChain(bytes.NewReader(self), int64(len(self))); !errors.Is(err, ErrXRefLoop) {
		t.Fatalf("self loop: %v", err)
	}
	// A -> B -> A.
	a := withPrev(orig)
	ab := pdfxtest.AppendUpdate(a, nil, "")
	ab = setPrev(ab, len(a))
	if _, err := XRefChain(bytes.NewReader(ab), int64(len(ab))); !errors.Is(err, ErrXRefLoop) {
		t.Fatalf("A-B-A loop: %v", err)
	}
}

func TestXRefChainGarbageOffset(t *testing.T) {
	good := doc(nil)
	i := bytes.LastIndex(good, []byte("startxref\n"))
	for _, off := range []string{"12", "0", "40"} {
		b := append(bytes.Clone(good[:i]), "startxref\n"+off+"\n%%EOF\n"...)
		if _, err := XRefChain(bytes.NewReader(b), int64(len(b))); !errors.Is(err, ErrBadSection) {
			t.Errorf("startxref %s: %v", off, err)
		}
	}
	far := append(bytes.Clone(good[:i]), "startxref\n99999999\n%%EOF\n"...)
	if _, err := XRefChain(bytes.NewReader(far), int64(len(far))); !errors.Is(err, ErrBadSection) {
		t.Errorf("past the end: %v", err)
	}
}

func TestXRefChainNoStartXRef(t *testing.T) {
	for _, b := range [][]byte{nil, []byte("%PDF-1.7\nsample\n"), append(doc(nil), bytes.Repeat([]byte{0}, 2000)...)} {
		if _, err := XRefChain(bytes.NewReader(b), int64(len(b))); !errors.Is(err, ErrNoStartXRef) {
			t.Errorf("%d bytes: %v", len(b), err)
		}
	}
}

func TestXRefChainTrailingGarbage(t *testing.T) {
	b := append(doc(nil), bytes.Repeat([]byte{0, 'x'}, 300)...)
	if c := chainOf(t, b); len(c) != 1 {
		t.Fatalf("chain %+v", c)
	}
}

func TestXRefChainLinearizedNoUpdate(t *testing.T) {
	b := doc(func(d *pdfxtest.Doc) { d.Linearized = true })
	c := chainOf(t, b)
	if !linearized(t, b) || len(c) != 2 || CountRevisions(c, true) != 1 {
		t.Fatalf("linearized %v chain %+v", linearized(t, b), c)
	}
}

func TestXRefChainLinearizedUpdate(t *testing.T) {
	b := update(doc(func(d *pdfxtest.Doc) { d.Linearized = true }))
	c := chainOf(t, b)
	if !linearized(t, b) || len(c) != 3 || CountRevisions(c, true) != 2 {
		t.Fatalf("chain %+v", c)
	}
}

func TestIsLinearizedFalseWhenLate(t *testing.T) {
	b := doc(func(d *pdfxtest.Doc) { d.Linearized = true })
	nl := bytes.IndexByte(b, '\n')
	late := append(append(bytes.Clone(b[:nl+1]), "%"+strings.Repeat("x", 2000)+"\n"...), b[nl+1:]...)
	if linearized(t, late) {
		t.Fatal("linearization dict after byte 1024 counted")
	}
	if linearized(t, doc(nil)) {
		t.Fatal("plain file counted as linearized")
	}
}

func TestCountRevisions(t *testing.T) {
	one := []Section{{Offset: 10}}
	two := []Section{{Offset: 20, HasPrev: true, Prev: 10}, {Offset: 10}}
	if CountRevisions(one, false) != 1 || CountRevisions(two, false) != 2 || CountRevisions(two, true) != 1 || CountRevisions(one, true) != 1 {
		t.Fatal("CountRevisions")
	}
}

func TestTokenizerStringsAndNesting(t *testing.T) {
	// Keys inside strings, nested dicts and arrays are not top-level keys.
	d := []byte(`<< /A (a /Prev 5 \) still a string) /B << /Prev 7 >> /C [ /Prev 8 ] /D <414243> /Prev 9 /XRefStm 3 /E 1 0 R >>`)
	keys, err := topLevel(d)
	if err != nil {
		t.Fatal(err)
	}
	if keys["Prev"] != "9" {
		t.Fatalf("Prev = %q", keys["Prev"])
	}
	if _, ok := keys["XRefStm"]; !ok {
		t.Fatal("XRefStm missing")
	}
	if _, err := topLevel([]byte("<< /Prev 1 ")); err == nil {
		t.Fatal("unterminated dict accepted")
	}
}
