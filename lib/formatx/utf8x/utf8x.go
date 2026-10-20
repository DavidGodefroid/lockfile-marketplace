// Package utf8x validates UTF-8 while streaming, so a plugin never holds
// the whole file to check it.
package utf8x

import (
	"errors"
	"io"
	"unicode/utf8"
)

// ErrInvalid is returned by a Reader on the first invalid sequence.
var ErrInvalid = errors.New("utf8x: invalid UTF-8")

// Valid reports whether everything r yields is valid UTF-8. An error is a
// read error, never a verdict.
func Valid(r io.Reader) (bool, error) {
	_, err := io.Copy(io.Discard, NewReader(r))
	if errors.Is(err, ErrInvalid) {
		return false, nil
	}
	return err == nil, err
}

// Reader passes r through and fails with ErrInvalid at the first invalid
// UTF-8 sequence, including one cut by the end of the input.
type Reader struct {
	r    io.Reader
	tail []byte // an incomplete sequence held back from the last read
	err  error
}

// NewReader returns a validating Reader over r.
func NewReader(r io.Reader) *Reader { return &Reader{r: r} }

func (v *Reader) Read(p []byte) (int, error) {
	if v.err != nil {
		return 0, v.err
	}
	if len(p) < utf8.UTFMax {
		return 0, io.ErrShortBuffer
	}
	n := copy(p, v.tail)
	m, err := v.r.Read(p[n:])
	n += m
	v.tail = v.tail[:0]
	if err != nil && !errors.Is(err, io.EOF) {
		v.err = err
		return 0, err
	}
	end := n
	if err == nil {
		end = completeUpTo(p[:n])
		v.tail = append(v.tail, p[end:n]...)
	}
	if !utf8.Valid(p[:end]) {
		v.err = ErrInvalid
		return 0, ErrInvalid
	}
	if errors.Is(err, io.EOF) {
		v.err = io.EOF
	}
	return end, v.err
}

// completeUpTo returns the length of b without a trailing sequence that
// may still be completed by the next read.
func completeUpTo(b []byte) int {
	for i := 1; i < utf8.UTFMax && i <= len(b); i++ {
		c := b[len(b)-i]
		if c < 0x80 {
			return len(b)
		}
		if c >= 0xC0 { // a leading byte: is its sequence complete?
			need := 2
			switch {
			case c >= 0xF0:
				need = 4
			case c >= 0xE0:
				need = 3
			}
			if i < need {
				return len(b) - i
			}
			return len(b)
		}
	}
	return len(b)
}
