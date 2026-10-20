package pdfx

import (
	"errors"
	"io"
	"math"
	"strconv"
)

// ErrBadLinearization is returned when a linearization dictionary has no
// valid /L.
var ErrBadLinearization = errors.New("pdfx: linearization dictionary without a valid /L")

// Linearization reports whether the first indirect object in the first
// 1024 bytes is a linearization dictionary (same rule as IsLinearized) and
// returns its /L, the file length it was written for.
func Linearization(r io.ReaderAt, size int64) (ok bool, length int64, err error) {
	head, err := readAt(r, size, 0, headBytes)
	if err != nil {
		return false, 0, err
	}
	m := firstObjRe.FindIndex(head)
	if m == nil {
		return false, 0, nil
	}
	b, err := readAt(r, size, int64(m[1]), maxDictBytes)
	if err != nil {
		return false, 0, err
	}
	keys, err := topLevel(b)
	if err != nil {
		return false, 0, nil
	}
	if _, ok := keys["Linearized"]; !ok {
		return false, 0, nil
	}
	l, err := strconv.ParseInt(keys["L"], 10, 64)
	if err != nil || l < 0 {
		return false, 0, ErrBadLinearization
	}
	return true, l, nil
}

// SameSize reports whether a and b have the same size within 1 mm on each
// side, orientation ignored.
func SameSize(a, b Rect) bool {
	aw, ah := a.Size()
	bw, bh := b.Size()
	return math.Abs(min(aw, ah)-min(bw, bh)) <= A4Tolerance &&
		math.Abs(max(aw, ah)-max(bw, bh)) <= A4Tolerance
}

// Orientation tells how a page with the visible box and the normalised
// /Rotate (0, 90, 180 or 270) is displayed. A page whose sides differ by
// 1 mm or less is square: neither portrait nor landscape.
func Orientation(box Rect, rotate int) (portrait, landscape bool) {
	w, h := box.Size()
	if rotate == 90 || rotate == 270 {
		w, h = h, w
	}
	if math.Abs(w-h) <= A4Tolerance {
		return false, false
	}
	return h > w, w > h
}
