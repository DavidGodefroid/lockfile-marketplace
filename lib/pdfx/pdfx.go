// Package pdfx holds the PDF logic shared by the marketplace's PDF
// plugins, on top of pdfcpu. Every function answers a structural question
// and returns errors that carry no content of the file.
package pdfx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

var (
	// ErrParse is returned when pdfcpu panics on the input.
	ErrParse = errors.New("pdfx: cannot parse the file")
	// ErrNotPDF is returned when the file has no PDF header.
	ErrNotPDF = errors.New("pdfx: not a PDF file")
	// ErrNoPages is returned by questions about pages on a file without any.
	ErrNoPages = errors.New("pdfx: no pages")
)

// HeaderWindow is how far into the file the %PDF- header may start.
const HeaderWindow = 1024

func init() {
	// No pdfcpu configuration directory: the plugin sandbox has no
	// filesystem, and pdfcpu would try to create one on first use.
	model.ConfigPath = "disable"
}

func config() *model.Configuration {
	c := model.NewStatelessConfiguration()
	c.ValidationMode = model.ValidationRelaxed
	c.ValidateLinks = false
	c.Offline = true
	return c
}

func catch(err *error) {
	if p := recover(); p != nil {
		*err = ErrParse
	}
}

// Open reads and validates a PDF in pdfcpu's relaxed mode.
func Open(r io.ReadSeeker) (ctx *model.Context, err error) {
	defer catch(&err)
	ctx, err = api.ReadContext(context.Background(), r, config())
	if err != nil {
		return nil, err
	}
	if err := api.ValidateContext(context.Background(), ctx); err != nil {
		return nil, err
	}
	return ctx, nil
}

// Encrypted reports whether the PDF has an /Encrypt dictionary. A file
// that needs a user password counts as encrypted: pdfcpu only asks for one
// after finding the dictionary.
func Encrypted(r io.ReadSeeker) (enc bool, err error) {
	defer catch(&err)
	ctx, err := api.ReadContext(context.Background(), r, config())
	if errors.Is(err, pdfcpu.ErrWrongPassword) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return ctx.Encrypt != nil, nil
}

// HasHeader reports whether %PDF- appears within the first HeaderWindow
// bytes.
func HasHeader(r io.Reader) (bool, error) {
	b := make([]byte, HeaderWindow)
	n, err := io.ReadFull(r, b)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return false, err
	}
	return bytes.Contains(b[:n], []byte("%PDF-")), nil
}

// TailWindow is how close to the end %%EOF must be.
const TailWindow = 1024

// LooksTruncated reports whether a PDF has no %%EOF in its last
// TailWindow bytes, or a last startxref that does not point at a cross
// reference section ("xref" or "N G obj"). A file without a PDF header is
// an error.
func LooksTruncated(r io.ReaderAt, size int64) (bool, error) {
	ok, err := HasHeader(io.NewSectionReader(r, 0, size))
	if err != nil {
		return false, err
	}
	if !ok {
		return false, ErrNotPDF
	}
	start := max(size-TailWindow, 0)
	tail := make([]byte, size-start)
	if _, err := r.ReadAt(tail, start); err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	eof := bytes.LastIndex(tail, []byte("%%EOF"))
	if eof < 0 {
		return true, nil
	}
	sx := bytes.LastIndex(tail[:eof], []byte("startxref"))
	if sx < 0 {
		return true, nil
	}
	var off int64
	if _, err := fmt.Sscan(string(tail[sx+len("startxref"):eof]), &off); err != nil {
		return true, nil
	}
	if off < 0 || off >= size {
		return true, nil
	}
	head := make([]byte, 64)
	n, err := r.ReadAt(head, off)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	return !isXRefStart(head[:n]), nil
}

func isXRefStart(b []byte) bool {
	if bytes.HasPrefix(b, []byte("xref")) {
		return true
	}
	// An xref stream: "<num> <gen> obj".
	var num, gen int
	var kw string
	if _, err := fmt.Sscan(string(b), &num, &gen, &kw); err != nil {
		return false
	}
	return num > 0 && gen >= 0 && kw == "obj"
}
