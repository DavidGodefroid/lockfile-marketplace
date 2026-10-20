package pdfx

import (
	"bytes"
	"errors"
	"regexp"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func encryptDict(b []byte) (types.Dict, error) {
	return EncryptDict(bytes.NewReader(b), int64(len(b)))
}

func mustEncryptDict(t *testing.T, b []byte) types.Dict {
	t.Helper()
	d, err := encryptDict(b)
	if err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("no /Encrypt dictionary")
	}
	return d
}

func intEntry(t *testing.T, d types.Dict, key string) int {
	t.Helper()
	v, ok := d[key].(types.Integer)
	if !ok {
		t.Fatalf("/%s = %v, want an integer", key, d[key])
	}
	return v.Value()
}

const stdV4 = "<< /Filter /Standard /V 4 /R 4 /Length 128 /P -4 /O (sample) /U (sample)" +
	" /CF << /StdCF << /CFM /V2 /Length 16 >> >> /StmF /StdCF /StrF /StdCF >>"

func TestEncryptDictPlain(t *testing.T) {
	for name, d := range map[string]pdfxtest.Doc{
		"table":  pdfxtest.Pages("text"),
		"stream": {Pages: []pdfxtest.Page{{}}, XRefStream: true},
		"objstm": {Pages: []pdfxtest.Page{{}}, ObjStream: true},
	} {
		got, err := encryptDict(pdfxtest.Write(d))
		if err != nil || got != nil {
			t.Errorf("%s: got %v, %v; want nil, nil", name, got, err)
		}
	}
}

func TestEncryptDictPdfcpu(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	for _, c := range []struct {
		name    string
		aes     bool
		bits    int
		v, r    int
		methods map[string]string
	}{
		{"rc4-40", false, 40, 1, 2, map[string]string{}},
		{"rc4-128", false, 128, 4, 4, map[string]string{"StdCF": "V2"}},
		{"aes-128", true, 128, 4, 4, map[string]string{"StdCF": "AESV2"}},
		{"aes-256", true, 256, 5, 5, map[string]string{"StdCF": "AESV3"}},
	} {
		// A user password: pdfcpu cannot open the file, EncryptDict can.
		d := mustEncryptDict(t, pdfxtest.EncryptCipher(t, plain, "sample-user", "sample-owner", c.aes, c.bits))
		if v, r := intEntry(t, d, "V"), intEntry(t, d, "R"); v != c.v || r != c.r {
			t.Errorf("%s: V %d R %d, want %d %d", c.name, v, r, c.v, c.r)
		}
		if f, _ := d["Filter"].(types.Name); f != "Standard" {
			t.Errorf("%s: Filter %v", c.name, d["Filter"])
		}
		m, err := CryptFilterMethods(d)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if len(m) != len(c.methods) {
			t.Errorf("%s: methods %v, want %v", c.name, m, c.methods)
		}
		for k, v := range c.methods {
			if m[k] != v {
				t.Errorf("%s: methods %v, want %v", c.name, m, c.methods)
			}
		}
	}
}

func TestEncryptDictLayouts(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	stream := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, XRefStream: true})
	objstm := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, ObjStream: true})
	cf := "<< /StdCF {{obj 1}} >>"
	nestedV4 := "<< /Filter /Standard /V {{obj 2}} /R 4 /CF " + cf + " /StmF /StdCF /StrF /StdCF >>"
	for name, b := range map[string][]byte{
		"direct":          pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+stdV4),
		"indirect table":  pdfxtest.AppendUpdate(plain, []string{stdV4}, "/Encrypt {{obj 0}}"),
		"indirect stream": pdfxtest.AppendUpdate(stream, []string{stdV4}, "/Encrypt {{obj 0}}"),
		"indirect objstm": pdfxtest.AppendUpdate(objstm, []string{stdV4}, "/Encrypt {{obj 0}}"),
		"nested refs": pdfxtest.AppendUpdate(plain,
			[]string{nestedV4, "<< /CFM /V2 >>", "4"}, "/Encrypt {{obj 0}}"),
		"older trailer": pdfxtest.AppendUpdate(
			pdfxtest.AppendUpdate(plain, []string{stdV4}, "/Encrypt {{obj 0}}"), nil, ""),
		"older section": pdfxtest.AppendUpdate(
			pdfxtest.AppendUpdate(stream, []string{stdV4}, ""), nil, "/Encrypt 5 0 R"),
	} {
		d, err := encryptDict(b)
		if err != nil || d == nil {
			t.Errorf("%s: %v, %v", name, d, err)
			continue
		}
		if v, ok := d["V"].(types.Integer); !ok || v != 4 {
			t.Errorf("%s: V %v", name, d["V"])
		}
		m, err := CryptFilterMethods(d)
		if err != nil || m["StdCF"] != "V2" {
			t.Errorf("%s: methods %v, %v", name, m, err)
		}
	}
}

var encryptRefRe = regexp.MustCompile(`/Encrypt (\d+) 0 R`)

// wrongGen makes the trailer's /Encrypt reference name generation 1.
func wrongGen(b []byte) []byte {
	return encryptRefRe.ReplaceAll(b, []byte("/Encrypt $1 1 R"))
}

func TestEncryptDictErrors(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	objstm := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, Objects: []string{stdV4}, ObjStream: true})
	for name, c := range map[string]struct {
		b    []byte
		want error
	}{
		"not a pdf":     {pdfxtest.Sample, ErrNotPDF},
		"no startxref":  {[]byte("%PDF-1.7\nsample\n"), ErrNoStartXRef},
		"integer":       {pdfxtest.AppendUpdate(plain, nil, "/Encrypt 5"), ErrBadEncrypt},
		"missing obj":   {pdfxtest.AppendUpdate(plain, nil, "/Encrypt 999 0 R"), ErrBadEncrypt},
		"wrong gen":     {wrongGen(pdfxtest.AppendUpdate(plain, []string{stdV4}, "/Encrypt {{obj 0}}")), ErrBadEncrypt},
		"not a dict":    {pdfxtest.AppendUpdate(plain, []string{"[1 2]"}, "/Encrypt {{obj 0}}"), ErrBadEncrypt},
		"cycle":         {pdfxtest.AppendUpdate(plain, []string{"<< /Filter /Standard /CF {{obj 0}} >>"}, "/Encrypt {{obj 0}}"), ErrBadEncrypt},
		"in objstm":     {pdfxtest.AppendUpdate(objstm, nil, "/Encrypt 4 0 R"), ErrBadEncrypt},
		"free entry":    {pdfxtest.AppendUpdate(plain, nil, "/Encrypt 0 0 R"), ErrBadEncrypt},
		"bad CF":        {pdfxtest.AppendUpdate(plain, nil, "/Encrypt << /CF 1 >>"), nil},
		"bad CF filter": {pdfxtest.AppendUpdate(plain, nil, "/Encrypt << /CF << /StdCF /V2 >> >>"), nil},
	} {
		d, err := encryptDict(c.b)
		if c.want == nil && err == nil {
			_, err = CryptFilterMethods(d)
			c.want = ErrBadEncrypt
		}
		if !errors.Is(err, c.want) {
			t.Errorf("%s: err %v, want %v", name, err, c.want)
		}
	}
}
