package pdfx

import (
	"errors"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ErrBadByteRange is returned for a /ByteRange entry that is not an
// integer.
var ErrBadByteRange = errors.New("pdfx: /ByteRange entry not an integer")

// CoversWholeFile reports whether the /ByteRange of the signature
// dictionary sig is [0 b c d] with b < c and c+d == size: the signed bytes
// run from the start of the file to its end, skipping one gap (the
// /Contents string). A byte range with more gaps does not cover the file;
// one that is not an array of integers is an error. Nothing is verified.
func CoversWholeFile(ctx *model.Context, sig types.Dict, size int64) (covers bool, err error) {
	defer catch(&err)
	o, ok := sig.Find("ByteRange")
	if !ok {
		return false, nil
	}
	br, err := ctx.DereferenceArray(o)
	if err != nil || len(br) != 4 {
		return false, err
	}
	var n [4]int64
	for i, e := range br {
		e, err := ctx.Dereference(e)
		if err != nil {
			return false, err
		}
		v, ok := e.(types.Integer)
		if !ok {
			return false, ErrBadByteRange
		}
		n[i] = int64(v.Value())
	}
	return n[0] == 0 && n[1] < n[2] && n[3] >= 0 && n[2]+n[3] == size, nil
}
