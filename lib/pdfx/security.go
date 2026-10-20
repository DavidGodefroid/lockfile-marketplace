package pdfx

import (
	"context"
	"errors"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// User access permission bits of the /P entry (PDF 32000-1 table 22,
// bit n is 1 << (n-1)).
const (
	PermPrint = 1 << 2 // bit 3: print
	PermCopy  = 1 << 4 // bit 5: copy or extract text and graphics
)

// NeedsPassword reports whether the PDF cannot be opened without a user
// password: it has an /Encrypt dictionary and the empty user password
// fails. A file encrypted with an owner password only opens without one.
func NeedsPassword(r io.ReadSeeker) (need bool, err error) {
	defer catch(&err)
	_, err = api.ReadContext(context.Background(), r, config())
	if errors.Is(err, pdfcpu.ErrWrongPassword) {
		return true, nil
	}
	return false, err
}

// Permissions returns the /P entry of the PDF's /Encrypt dictionary, or
// encrypted false when there is none. A file that needs a user password is
// an error: its permissions cannot be read without opening it.
func Permissions(r io.ReadSeeker) (p int, encrypted bool, err error) {
	defer catch(&err)
	ctx, err := api.ReadContext(context.Background(), r, config())
	if err != nil {
		return 0, false, err
	}
	if ctx.Encrypt == nil || ctx.E == nil {
		return 0, false, nil
	}
	return ctx.E.P, true, nil
}
