package pdfx

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
)

var (
	ErrNoStartXRef = errors.New("pdfx: no startxref near the end of the file")
	ErrBadSection  = errors.New("pdfx: startxref or /Prev does not point at a cross-reference section")
	ErrXRefLoop    = errors.New("pdfx: cross-reference chain loops or is too long")
)

// Bounds of the chain walk.
const (
	maxSections  = 1024
	maxDictBytes = 64 << 10
	tailBytes    = 1024
	headBytes    = 1024
)

// Section is one cross-reference section of the chain.
type Section struct {
	Offset     int64
	Stream     bool // an xref stream, not a classic table
	Prev       int64
	HasPrev    bool
	HasXRefStm bool // hybrid trailer
}

// XRefChain walks the cross-reference chain: the last startxref in the
// final 1024 bytes, then each /Prev. It reads only section headers and
// trailer or xref-stream dictionaries, never stream data, with its own
// tokenizer (pdfcpu repairs chains silently).
func XRefChain(r io.ReaderAt, size int64) ([]Section, error) {
	off, err := lastStartXRef(r, size)
	if err != nil {
		return nil, err
	}
	seen := map[int64]bool{}
	var chain []Section
	for {
		if seen[off] || len(chain) >= maxSections {
			return nil, ErrXRefLoop
		}
		seen[off] = true
		s, err := readSection(r, size, off)
		if err != nil {
			return nil, err
		}
		chain = append(chain, s)
		if !s.HasPrev {
			return chain, nil
		}
		off = s.Prev
	}
}

// CountRevisions is the number of saves the chain records: one per section,
// minus the first-page section of a linearized file (the chain then has a
// section with /Prev that is not an update).
func CountRevisions(chain []Section, linearized bool) int {
	n := len(chain)
	if linearized {
		for _, s := range chain {
			if s.HasPrev {
				return n - 1
			}
		}
	}
	return n
}

var firstObjRe = regexp.MustCompile(`(\d+)\s+(\d+)\s+obj\b`)

// IsLinearized reports whether the first indirect object starts within the
// first 1024 bytes and its dictionary has /Linearized.
func IsLinearized(r io.ReaderAt, size int64) (bool, error) {
	head, err := readAt(r, size, 0, headBytes)
	if err != nil {
		return false, err
	}
	m := firstObjRe.FindIndex(head)
	if m == nil {
		return false, nil
	}
	b, err := readAt(r, size, int64(m[1]), maxDictBytes)
	if err != nil {
		return false, err
	}
	keys, err := topLevel(b)
	if err != nil {
		return false, nil
	}
	_, ok := keys["Linearized"]
	return ok, nil
}

func readAt(r io.ReaderAt, size, off int64, n int) ([]byte, error) {
	if off < 0 || off >= size {
		return nil, nil
	}
	n = int(min(int64(n), size-off))
	b := make([]byte, n)
	got, err := r.ReadAt(b, off)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return b[:got], nil
}

func lastStartXRef(r io.ReaderAt, size int64) (int64, error) {
	start := max(size-tailBytes, 0)
	tail, err := readAt(r, size, start, tailBytes)
	if err != nil {
		return 0, err
	}
	i := bytes.LastIndex(tail, []byte("startxref"))
	if i < 0 {
		return 0, ErrNoStartXRef
	}
	rest := bytes.TrimLeft(tail[i+len("startxref"):], " \t\r\n\f\x00")
	j := 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	if j == 0 || j > 18 {
		return 0, ErrNoStartXRef
	}
	off, _ := strconv.ParseInt(string(rest[:j]), 10, 64)
	return off, nil
}

func isWS(c byte) bool { return c == 0 || c == '\t' || c == '\n' || c == '\f' || c == '\r' || c == ' ' }

var objHeadRe = regexp.MustCompile(`^(\d+)\s+(\d+)\s+obj\b`)

func readSection(r io.ReaderAt, size, off int64) (Section, error) {
	if off < 0 || off >= size {
		return Section{}, ErrBadSection
	}
	s := Section{Offset: off}
	head, err := readAt(r, size, off, maxDictBytes)
	if err != nil {
		return Section{}, err
	}
	var dict []byte
	switch {
	case bytes.HasPrefix(head, []byte("xref")):
		t, err := findTrailer(r, size, off+4)
		if err != nil {
			return Section{}, err
		}
		if dict, err = readAt(r, size, t, maxDictBytes); err != nil {
			return Section{}, err
		}
	case objHeadRe.Match(head):
		s.Stream = true
		dict = head[objHeadRe.FindIndex(head)[1]:]
	default:
		return Section{}, ErrBadSection
	}
	keys, err := topLevel(dict)
	if err != nil {
		return Section{}, ErrBadSection
	}
	if s.Stream && keys["Type"] != "/XRef" {
		return Section{}, ErrBadSection
	}
	if v, ok := keys["Prev"]; ok {
		p, err := strconv.ParseInt(v, 10, 64)
		if err != nil || p < 0 {
			return Section{}, ErrBadSection
		}
		s.Prev, s.HasPrev = p, true
	}
	_, s.HasXRefStm = keys["XRefStm"]
	return s, nil
}

// findTrailer scans a classic table from off for the trailer keyword and
// returns the offset just after it. Anything but entry characters before it
// is not a table.
func findTrailer(r io.ReaderAt, size, off int64) (int64, error) {
	const chunk = 64 << 10
	kw := []byte("trailer")
	for pos := off; pos < size; pos += chunk {
		// Each read overlaps the next chunk by the keyword's length, so a
		// keyword starting in this chunk is read whole.
		b, err := readAt(r, size, pos, chunk+len(kw))
		if err != nil {
			return 0, err
		}
		for i := 0; i < min(len(b), chunk); i++ {
			c := b[i]
			if c == 't' {
				if bytes.HasPrefix(b[i:], kw) {
					return pos + int64(i+len(kw)), nil
				}
				return 0, ErrBadSection
			}
			if !(isWS(c) || (c >= '0' && c <= '9') || c == 'f' || c == 'n') {
				return 0, ErrBadSection
			}
		}
	}
	return 0, ErrBadSection
}

// topLevel parses the dictionary at the start of b (after whitespace and
// comments) and returns each top-level key with the text of its value's
// first token: a name with its slash, a number, a keyword; nested values
// are skipped, strings never kept.
func topLevel(b []byte) (map[string]string, error) {
	l := lexer{b: b}
	if l.next() != "<<" {
		return nil, ErrBadSection
	}
	keys := map[string]string{}
	for {
		t := l.next()
		switch {
		case t == ">>":
			return keys, nil
		case t == "":
			return nil, ErrBadSection
		case t[0] != '/':
			return nil, ErrBadSection
		}
		v, err := l.value()
		if err != nil {
			return nil, err
		}
		keys[t[1:]] = v
	}
}

type lexer struct {
	b []byte
	i int
}

const (
	tokString = "(string)"
	tokHex    = "<hex>"
)

// next returns the next token: "<<", ">>", "[", "]", a name with its slash,
// a number or keyword, tokString or tokHex; "" at the end or on an
// unterminated token.
func (l *lexer) next() string {
	for l.i < len(l.b) {
		c := l.b[l.i]
		switch {
		case isWS(c):
			l.i++
			continue
		case c == '%':
			for l.i < len(l.b) && l.b[l.i] != '\n' && l.b[l.i] != '\r' {
				l.i++
			}
			continue
		}
		break
	}
	if l.i >= len(l.b) {
		return ""
	}
	c := l.b[l.i]
	switch {
	case bytes.HasPrefix(l.b[l.i:], []byte("<<")):
		l.i += 2
		return "<<"
	case bytes.HasPrefix(l.b[l.i:], []byte(">>")):
		l.i += 2
		return ">>"
	case c == '[' || c == ']':
		l.i++
		return string(c)
	case c == '(':
		depth := 0
		for ; l.i < len(l.b); l.i++ {
			switch l.b[l.i] {
			case '\\':
				l.i++
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					l.i++
					return tokString
				}
			}
		}
		return ""
	case c == '<':
		end := bytes.IndexByte(l.b[l.i:], '>')
		if end < 0 {
			return ""
		}
		l.i += end + 1
		return tokHex
	}
	start := l.i
	if c == '/' {
		l.i++
	}
	for l.i < len(l.b) && !isWS(l.b[l.i]) && !bytes.ContainsRune([]byte("()<>[]{}/%"), rune(l.b[l.i])) {
		l.i++
	}
	if l.i == start {
		l.i++
		return string(c)
	}
	return string(l.b[start:l.i])
}

// value consumes one value and returns its first token. "N G R" is one
// value.
func (l *lexer) value() (string, error) {
	t := l.next()
	switch t {
	case "", ">>", "]":
		return "", ErrBadSection
	case "<<":
		return t, l.skip(">>")
	case "[":
		return t, l.skip("]")
	}
	if isInt(t) {
		save := l.i
		if g := l.next(); isInt(g) {
			if l.next() == "R" {
				return t, nil
			}
		}
		l.i = save
	}
	return t, nil
}

// skip consumes tokens up to the close matching an already consumed open.
func (l *lexer) skip(close string) error {
	for depth := 0; depth < MaxDepth; {
		t := l.next()
		switch t {
		case "":
			return ErrBadSection
		case "<<", "[":
			depth++
		case ">>", "]":
			if depth == 0 {
				if t != close {
					return ErrBadSection
				}
				return nil
			}
			depth--
		}
	}
	return ErrBadSection
}

func isInt(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
