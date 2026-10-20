package pdfx

import (
	"bytes"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func ints(v ...int) types.Array {
	a := types.Array{}
	for _, n := range v {
		a = append(a, types.Integer(n))
	}
	return a
}

func TestCoversWholeFile(t *testing.T) {
	signed := pdfxtest.Sign(pdfxtest.Write(pdfxtest.Doc{
		Catalog: map[string]string{"/AcroForm": "<< /Fields [{{obj 0}}] /SigFlags 3 >>"},
		Objects: []string{pdfxtest.SignatureField(0), pdfxtest.SignatureValue()},
	}))
	ctx, err := Open(bytes.NewReader(signed))
	if err != nil {
		t.Fatal(err)
	}
	sigs, err := SignatureValues(ctx)
	if err != nil || len(sigs) != 1 {
		t.Fatalf("SignatureValues: %v, %v", sigs, err)
	}
	size := int64(len(signed))
	cases := []struct {
		name   string
		br     types.Object
		size   int64
		covers bool
		bad    bool
	}{
		{name: "signed", size: size, covers: true},
		{name: "longer file", size: size + 1},
		{name: "shorter file", size: size - 1},
		{name: "offset start", br: ints(1, 10, 20, 30), size: 50},
		{name: "empty gap", br: ints(0, 20, 20, 30), size: 50},
		{name: "reversed gap", br: ints(0, 30, 20, 30), size: 50},
		{name: "exact", br: ints(0, 10, 20, 30), size: 50, covers: true},
		{name: "two gaps", br: ints(0, 10, 20, 5, 30, 20), size: 50},
		{name: "real number", br: types.Array{types.Integer(0), types.Integer(10), types.Integer(20), types.Float(30)}, size: 50, bad: true},
		{name: "not an array", br: types.StringLiteral("sample"), size: 50, bad: true},
	}
	for _, c := range cases {
		sig := sigs[0].Clone().(types.Dict)
		if c.br != nil {
			sig["ByteRange"] = c.br
		}
		got, err := CoversWholeFile(ctx, sig, c.size)
		if (err != nil) != c.bad {
			t.Errorf("%s: err = %v, want error %v", c.name, err, c.bad)
		} else if got != c.covers {
			t.Errorf("%s: covers = %v, want %v", c.name, got, c.covers)
		}
	}
}
