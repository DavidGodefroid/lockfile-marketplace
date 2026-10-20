// Command pdf.uses-aes256-encryption answers: is the PDF encrypted with
// AES-256 as PDF 2.0 defines it?
package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		st, err := f.Stat()
		if err != nil {
			return false, err
		}
		d, err := pdfx.EncryptDict(f, st.Size())
		if err != nil || d == nil {
			return false, err
		}
		return usesAES256(d)
	})
}

// usesAES256 reports whether /V is 5, /R is at least 6 (R 5 is the
// deprecated Adobe extension) and both /StmF and /StrF name a crypt filter
// with /CFM /AESV3. Absent /V, /R, /StmF or /StrF take their defaults
// (0, 0, Identity, Identity).
func usesAES256(d types.Dict) (bool, error) {
	v, err := integer(d, "V")
	if err != nil {
		return false, err
	}
	r, err := integer(d, "R")
	if err != nil {
		return false, err
	}
	methods, err := pdfx.CryptFilterMethods(d)
	if err != nil {
		return false, err
	}
	stm, err := method(d, "StmF", methods)
	if err != nil {
		return false, err
	}
	str, err := method(d, "StrF", methods)
	if err != nil {
		return false, err
	}
	return v == 5 && r >= 6 && stm == "AESV3" && str == "AESV3", nil
}

// integer returns d[key] as an integer, 0 when absent.
func integer(d types.Dict, key string) (int, error) {
	o, ok := d[key]
	if !ok {
		return 0, nil
	}
	n, ok := o.(types.Integer)
	if !ok {
		return 0, pdfx.ErrBadEncrypt
	}
	return n.Value(), nil
}

// method returns the /CFM of the crypt filter named by d[key]; Identity
// (also the default) has none. A name missing from /CF is an error.
func method(d types.Dict, key string, methods map[string]string) (string, error) {
	o, ok := d[key]
	if !ok {
		return "None", nil
	}
	n, ok := o.(types.Name)
	if !ok {
		return "", pdfx.ErrBadEncrypt
	}
	if n == "Identity" {
		return "None", nil
	}
	m, ok := methods[string(n)]
	if !ok {
		return "", pdfx.ErrBadEncrypt
	}
	return m, nil
}
