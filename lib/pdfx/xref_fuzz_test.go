package pdfx

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func FuzzXRefChain(f *testing.F) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	f.Add(plain)
	f.Add(pdfxtest.AppendUpdate(plain, []string{"<< /Producer (sample) >>"}, ""))
	for _, mod := range []func(*pdfxtest.Doc){
		func(d *pdfxtest.Doc) { d.XRefStream = true },
		func(d *pdfxtest.Doc) { d.Hybrid = true },
		func(d *pdfxtest.Doc) { d.Linearized = true },
	} {
		d := pdfxtest.Pages("text")
		mod(&d)
		f.Add(pdfxtest.Write(d))
	}
	f.Add([]byte("startxref\n0\n%%EOF"))
	f.Fuzz(func(t *testing.T, b []byte) {
		r := bytes.NewReader(b)
		c, err := XRefChain(r, int64(len(b)))
		if err == nil && len(c) == 0 {
			t.Fatal("no error and no section")
		}
		IsLinearized(r, int64(len(b)))
	})
}
