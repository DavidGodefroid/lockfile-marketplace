// Command xml.has-external-entity answers: does the XML document's DOCTYPE
// declare an external entity?
package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasExternalEntity(f)
	})
}

// errNotWellFormed is returned for input that is not a well-formed XML
// document: the question does not apply to it.
var errNotWellFormed = errors.New("not a well-formed XML document")

// hasExternalEntity reports whether the DOCTYPE of the XML document in r
// declares an entity, general or parameter, with a SYSTEM or PUBLIC
// identifier. The whole document must be well-formed: exactly one root
// element, only whitespace outside it, a single DOCTYPE before it and the
// XML declaration only as the very first token. References to entities the
// DOCTYPE declares are accepted. An error from r itself is returned as is.
// A leading UTF-8 byte order mark is skipped.
func hasExternalEntity(r io.Reader) (bool, error) {
	src := &recorder{r: r}
	br := bufio.NewReader(src)
	if bom, _ := br.Peek(3); bytes.Equal(bom, []byte("\xef\xbb\xbf")) {
		_, _ = br.Discard(3)
	}
	if src.err != nil {
		return false, src.err
	}
	d := xml.NewDecoder(br)
	d.Strict = true
	d.CharsetReader = charsetReader
	d.Entity = map[string]string{}
	depth, roots := 0, 0
	doctype, external := false, false
	for i := 0; ; i++ {
		tok, err := d.Token()
		if src.err != nil {
			return false, src.err
		}
		if errors.Is(err, io.EOF) {
			if depth != 0 || roots != 1 {
				return false, errNotWellFormed
			}
			return external, nil
		}
		if err != nil {
			return false, fmt.Errorf("%w: %v", errNotWellFormed, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots > 1 {
					return false, fmt.Errorf("%w: several root elements", errNotWellFormed)
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.Trim(t, " \t\r\n")) > 0 {
				return false, fmt.Errorf("%w: text outside the root element", errNotWellFormed)
			}
		case xml.Directive:
			if roots > 0 || doctype {
				return false, fmt.Errorf("%w: misplaced or repeated directive", errNotWellFormed)
			}
			doctype = true
			decls, err := entityDecls(t)
			if err != nil {
				return false, fmt.Errorf("%w: %v", errNotWellFormed, err)
			}
			for _, e := range decls {
				external = external || e.external
				if _, seen := d.Entity[e.name]; !seen && !e.parameter {
					d.Entity[e.name] = ""
				}
			}
		case xml.ProcInst:
			if t.Target == "xml" && i > 0 {
				return false, fmt.Errorf("%w: misplaced XML declaration", errNotWellFormed)
			}
		}
	}
}

// entityDecl is one <!ENTITY> declaration of the internal subset.
type entityDecl struct {
	name      string
	parameter bool // a parameter entity (<!ENTITY % name ...>)
	external  bool // declared with a SYSTEM or PUBLIC identifier
}

// entityDecls reads a directive body (the bytes between "<!" and ">") that
// must be a DOCTYPE, and returns the entity declarations of its internal
// subset. The decoder has already replaced comments with a space.
func entityDecls(dir []byte) ([]entityDecl, error) {
	s := &scanner{b: dir}
	if !s.keyword("DOCTYPE") {
		return nil, errors.New("directive is not a DOCTYPE")
	}
	// Root name and optional external identifier, up to the internal subset.
	for {
		s.space()
		if s.done() {
			return nil, nil
		}
		if s.peek() == '[' {
			s.i++
			break
		}
		if _, err := s.token(); err != nil {
			return nil, err
		}
	}
	var decls []entityDecl
	for {
		s.space()
		switch {
		case s.done():
			return nil, errors.New("unterminated internal subset")
		case s.peek() == ']':
			s.i++
			s.space()
			if !s.done() {
				return nil, errors.New("data after the internal subset")
			}
			return decls, nil
		case s.prefix("<?"):
			if !s.skipPast("?>") {
				return nil, errors.New("unterminated processing instruction")
			}
		case s.prefix("<!ENTITY"):
			s.i += len("<!ENTITY")
			e, err := s.entity()
			if err != nil {
				return nil, err
			}
			decls = append(decls, e)
		case s.prefix("<!"):
			if err := s.skipDecl(); err != nil {
				return nil, err
			}
		case s.peek() == '%':
			if !s.skipPast(";") {
				return nil, errors.New("unterminated parameter entity reference")
			}
		default:
			return nil, fmt.Errorf("unexpected byte %q in the internal subset", s.peek())
		}
	}
}

// scanner walks a directive body.
type scanner struct {
	b []byte
	i int
}

func (s *scanner) done() bool { return s.i >= len(s.b) }

func (s *scanner) peek() byte { return s.b[s.i] }

func (s *scanner) prefix(p string) bool { return bytes.HasPrefix(s.b[s.i:], []byte(p)) }

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }

// space skips whitespace and reports whether there was any.
func (s *scanner) space() bool {
	start := s.i
	for !s.done() && isSpace(s.peek()) {
		s.i++
	}
	return s.i > start
}

// keyword consumes kw when it is followed by whitespace.
func (s *scanner) keyword(kw string) bool {
	if !s.prefix(kw) || s.i+len(kw) >= len(s.b) || !isSpace(s.b[s.i+len(kw)]) {
		return false
	}
	s.i += len(kw)
	return true
}

// skipPast moves after the next occurrence of end.
func (s *scanner) skipPast(end string) bool {
	k := bytes.Index(s.b[s.i:], []byte(end))
	if k < 0 {
		return false
	}
	s.i += k + len(end)
	return true
}

// token consumes a quoted literal, returned with its quotes, or a run of
// bytes up to whitespace, '>' or '['.
func (s *scanner) token() (string, error) {
	start := s.i
	if q := s.peek(); q == '"' || q == '\'' {
		k := bytes.IndexByte(s.b[s.i+1:], q)
		if k < 0 {
			return "", errors.New("unterminated literal")
		}
		s.i += k + 2
		return string(s.b[start:s.i]), nil
	}
	for !s.done() && !isSpace(s.peek()) && s.peek() != '>' && s.peek() != '[' {
		s.i++
	}
	if s.i == start {
		return "", fmt.Errorf("unexpected byte %q", s.peek())
	}
	return string(s.b[start:s.i]), nil
}

func isLiteral(tok string) bool { return tok[0] == '"' || tok[0] == '\'' }

// entity parses the rest of an <!ENTITY declaration, after the keyword:
// ['%'] Name (EntityValue | (SYSTEM | PUBLIC) literals [NDATA Name]) '>'.
func (s *scanner) entity() (entityDecl, error) {
	var e entityDecl
	var toks []string
	for {
		if !s.space() && len(toks) == 0 {
			return e, errors.New("malformed ENTITY declaration")
		}
		if s.done() {
			return e, errors.New("unterminated ENTITY declaration")
		}
		if s.peek() == '>' {
			s.i++
			break
		}
		tok, err := s.token()
		if err != nil {
			return e, err
		}
		toks = append(toks, tok)
	}
	if len(toks) > 0 && toks[0] == "%" {
		e.parameter = true
		toks = toks[1:]
	}
	if len(toks) < 2 || isLiteral(toks[0]) {
		return e, errors.New("malformed ENTITY declaration")
	}
	e.name = toks[0]
	switch def := toks[1:]; {
	case len(def) == 1 && isLiteral(def[0]):
	case len(def) >= 2 && def[0] == "SYSTEM" && isLiteral(def[1]):
		e.external = true
		def = def[2:]
		if len(def) != 0 && !(len(def) == 2 && def[0] == "NDATA" && !e.parameter) {
			return e, errors.New("malformed external ENTITY declaration")
		}
	case len(def) >= 3 && def[0] == "PUBLIC" && isLiteral(def[1]) && isLiteral(def[2]):
		e.external = true
		def = def[3:]
		if len(def) != 0 && !(len(def) == 2 && def[0] == "NDATA" && !e.parameter) {
			return e, errors.New("malformed external ENTITY declaration")
		}
	default:
		return e, errors.New("malformed ENTITY declaration")
	}
	return e, nil
}

// skipDecl skips another markup declaration (ELEMENT, ATTLIST, NOTATION)
// up to its closing '>', ignoring '>' inside quoted literals.
func (s *scanner) skipDecl() error {
	var quote byte
	for s.i++; !s.done(); s.i++ {
		c := s.peek()
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			s.i++
			return nil
		}
	}
	return errors.New("unterminated markup declaration")
}

// recorder passes reads through and keeps the first error other than
// io.EOF, so that a failing read is told apart from malformed input.
type recorder struct {
	r   io.Reader
	err error
}

func (rc *recorder) Read(p []byte) (int, error) {
	n, err := rc.r.Read(p)
	if err != nil && !errors.Is(err, io.EOF) && rc.err == nil {
		rc.err = err
	}
	return n, err
}

// charsetReader accepts US-ASCII as is and converts ISO-8859-1 to UTF-8.
// Any other declared encoding fails the decoder.
func charsetReader(label string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(label) {
	case "us-ascii":
		return input, nil
	case "iso-8859-1", "latin1":
		return &latin1{r: input}, nil
	}
	return nil, fmt.Errorf("unsupported charset %q", label)
}

// latin1 decodes ISO-8859-1: every byte is the rune of the same value.
type latin1 struct {
	r   io.Reader
	buf [4096]byte
	out []byte
	err error
}

func (l *latin1) Read(p []byte) (int, error) {
	for len(l.out) == 0 {
		if l.err != nil {
			return 0, l.err
		}
		var n int
		n, l.err = l.r.Read(l.buf[:])
		for _, b := range l.buf[:n] {
			l.out = utf8.AppendRune(l.out, rune(b))
		}
	}
	n := copy(p, l.out)
	l.out = l.out[n:]
	return n, nil
}
