package pdfx

import (
	"bytes"
	"context"
	"errors"
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// coverTolerance is how far (in points) a white rectangle may fall short
// of the visible page area and still count as a full-page fill.
const coverTolerance = 1.0

type rect struct{ llx, lly, urx, ury float64 }

func normRect(x, y, w, h float64) rect {
	r := rect{x, y, x + w, y + h}
	if r.llx > r.urx {
		r.llx, r.urx = r.urx, r.llx
	}
	if r.lly > r.ury {
		r.lly, r.ury = r.ury, r.lly
	}
	return r
}

func (r rect) covers(b rect) bool {
	return r.llx <= b.llx+coverTolerance && r.lly <= b.lly+coverTolerance &&
		r.urx >= b.urx-coverTolerance && r.ury >= b.ury-coverTolerance
}

func fromRectangle(r *types.Rectangle) rect {
	return normRect(r.LL.X, r.LL.Y, r.Width(), r.Height())
}

// visibleBox is the media box clipped by the crop box, or false when the
// page has no media box.
func visibleBox(inh *model.InheritedPageAttrs) (rect, bool) {
	if inh == nil || inh.MediaBox == nil {
		return rect{}, false
	}
	v := fromRectangle(inh.MediaBox)
	if inh.CropBox != nil {
		c := fromRectangle(inh.CropBox)
		v = rect{max(v.llx, c.llx), max(v.lly, c.lly), min(v.urx, c.urx), min(v.ury, c.ury)}
	}
	return v, true
}

// PagePaints reports whether page (1-based) paints anything visible: a
// painting operator other than a fill, in white, of rectangles one of which
// covers the visible page area. Text in an invisible rendering mode (3 or
// 7) paints nothing; images, XObjects and shadings always paint.
// Annotations are not considered.
func PagePaints(ctx *model.Context, page int) (paints bool, err error) {
	defer catch(&err)
	d, _, inh, err := ctx.PageDict(context.Background(), page, false)
	if err != nil {
		return false, err
	}
	content, err := ctx.PageContent(d, page)
	if errors.Is(err, model.ErrNoContent) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	box, hasBox := visibleBox(inh)
	return scanPaints(content, box, hasBox), nil
}

type gstate struct {
	white    bool // non-stroking colour is white
	identity bool // CTM is the identity
	textMode int
}

func scanPaints(content []byte, box rect, hasBox bool) bool {
	gs := gstate{identity: true}
	var stack []gstate
	var rects []rect
	otherPath := false
	clearPath := func() { rects, otherPath = nil, false }
	paints := false
	lex(content, func(op string, args []token) bool {
		nums := numbers(args)
		switch op {
		case "q":
			stack = append(stack, gs)
		case "Q":
			if n := len(stack); n > 0 {
				gs, stack = stack[n-1], stack[:n-1]
			}
		case "cm":
			if len(nums) != 6 || nums[0] != 1 || nums[1] != 0 || nums[2] != 0 || nums[3] != 1 || nums[4] != 0 || nums[5] != 0 {
				gs.identity = false
			}
		case "g":
			gs.white = len(nums) == 1 && nums[0] == 1
		case "rg":
			gs.white = len(nums) == 3 && nums[0] == 1 && nums[1] == 1 && nums[2] == 1
		case "k":
			gs.white = len(nums) == 4 && nums[0] == 0 && nums[1] == 0 && nums[2] == 0 && nums[3] == 0
		case "cs", "sc", "scn":
			// Colour spaces other than the device ones cannot be read as
			// white here: count it as a colour that paints.
			gs.white = false
		case "Tr":
			if len(nums) == 1 {
				gs.textMode = int(nums[0])
			}
		case "re":
			if len(nums) == 4 {
				rects = append(rects, normRect(nums[0], nums[1], nums[2], nums[3]))
			} else {
				otherPath = true
			}
		case "m", "l", "c", "v", "y", "h":
			otherPath = true
		case "n":
			clearPath()
		case "f", "F", "f*":
			if !(gs.white && gs.identity && hasBox && !otherPath && anyCovers(rects, box)) {
				paints = true
			}
			clearPath()
		case "S", "s", "B", "B*", "b", "b*":
			paints = true
		case "Tj", "TJ", "'", "\"":
			if gs.textMode != 3 && gs.textMode != 7 {
				paints = true
			}
		case "Do", "sh", "BI":
			paints = true
		}
		return !paints
	})
	return paints
}

func anyCovers(rs []rect, box rect) bool {
	for _, r := range rs {
		if r.covers(box) {
			return true
		}
	}
	return false
}

func numbers(args []token) []float64 {
	out := make([]float64, 0, len(args))
	for _, a := range args {
		if a.kind != tokNumber {
			return nil
		}
		out = append(out, a.num)
	}
	return out
}

type tokKind int

const (
	tokNumber tokKind = iota
	tokOther          // name, string, array, dict, boolean, null
)

type token struct {
	kind tokKind
	num  float64
}

func isSpace(c byte) bool {
	return c == 0 || c == '\t' || c == '\n' || c == '\f' || c == '\r' || c == ' '
}

func isDelim(c byte) bool {
	switch c {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	}
	return false
}

// lex calls fn for each operator of a content stream with its operands,
// until fn returns false. Strings, arrays and dicts are single operands; an
// inline image (BI … ID data EI) is reported as BI.
func lex(b []byte, fn func(op string, args []token) bool) {
	var args []token
	i := 0
	for i < len(b) {
		c := b[i]
		switch {
		case isSpace(c):
			i++
		case c == '%':
			for i < len(b) && b[i] != '\n' && b[i] != '\r' {
				i++
			}
		case c == '(':
			i = skipString(b, i)
			args = append(args, token{kind: tokOther})
		case c == '<' && i+1 < len(b) && b[i+1] == '<':
			i = skipNested(b, i, "<<", ">>")
			args = append(args, token{kind: tokOther})
		case c == '<':
			for i < len(b) && b[i] != '>' {
				i++
			}
			i++
			args = append(args, token{kind: tokOther})
		case c == '[':
			i = skipNested(b, i, "[", "]")
			args = append(args, token{kind: tokOther})
		case c == '/':
			i++
			for i < len(b) && !isSpace(b[i]) && !isDelim(b[i]) {
				i++
			}
			args = append(args, token{kind: tokOther})
		case isDelim(c):
			// Stray ) > ] } : skip.
			i++
		default:
			j := i
			for j < len(b) && !isSpace(b[j]) && !isDelim(b[j]) {
				j++
			}
			word := string(b[i:j])
			i = j
			if n, err := strconv.ParseFloat(word, 64); err == nil && (word[0] == '-' || word[0] == '+' || word[0] == '.' || (word[0] >= '0' && word[0] <= '9')) {
				args = append(args, token{kind: tokNumber, num: n})
				continue
			}
			switch word {
			case "true", "false", "null":
				args = append(args, token{kind: tokOther})
				continue
			case "BI":
				i = skipInlineImage(b, i)
			}
			if !fn(word, args) {
				return
			}
			args = args[:0]
		}
	}
}

func skipString(b []byte, i int) int {
	depth := 0
	for i < len(b) {
		switch b[i] {
		case '\\':
			i++
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
		i++
	}
	return i
}

// skipNested skips a balanced open…close group starting at b[i], strings
// inside it included.
func skipNested(b []byte, i int, open, close string) int {
	depth := 0
	for i < len(b) {
		switch {
		case b[i] == '(':
			i = skipString(b, i)
			continue
		case bytes.HasPrefix(b[i:], []byte(open)):
			depth++
			i += len(open)
			continue
		case bytes.HasPrefix(b[i:], []byte(close)):
			depth--
			i += len(close)
			if depth == 0 {
				return i
			}
			continue
		}
		i++
	}
	return i
}

// skipInlineImage moves past "… ID <data> EI" after a BI operator.
func skipInlineImage(b []byte, i int) int {
	id := bytes.Index(b[i:], []byte("ID"))
	if id < 0 {
		return len(b)
	}
	i += id + 2
	for i+2 <= len(b) {
		if b[i] == 'E' && b[i+1] == 'I' && (i == 0 || isSpace(b[i-1])) && (i+2 == len(b) || isSpace(b[i+2])) {
			return i + 2
		}
		i++
	}
	return len(b)
}
