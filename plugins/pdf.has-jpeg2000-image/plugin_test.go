package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk/harness"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestConformance(t *testing.T) {
	writeFixtures(t)
	harness.Check(t, ".")
}

func TestNoAnswer(t *testing.T) {
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt": pdfxtest.Sample,
		"empty":      nil,
	})
}

// pdfcpu drops a stream whose /Filter is malformed when it loads the file,
// so the error path of jpx is checked on dicts built here.
func TestMalformedFilter(t *testing.T) {
	ctx, err := pdfx.Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}})))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []types.Object{types.Integer(5), types.Array{types.Name("FlateDecode"), types.Integer(5)}} {
		if _, err := jpx(ctx, types.Dict{"Filter": f}); err == nil {
			t.Errorf("/Filter %v: want an error", f)
		}
	}
}
