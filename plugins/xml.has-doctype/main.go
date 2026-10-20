// Command xml.has-doctype answers: does the XML document have a DOCTYPE
// declaration?
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
		return hasDoctype(f)
	})
}

// errNotXML reports input that is not a well-formed XML document.
var errNotXML = errors.New("not well-formed XML")

// hasDoctype reports whether r, a well-formed XML document, declares a
// DOCTYPE in its prolog. Well-formed means the same as in xml.well-formed
// (one root element, only whitespace outside it, the XML declaration only
// at the very start), plus: the only directive allowed is a single
// DOCTYPE before the root. Anything else is errNotXML; an error from r
// itself is returned as is. A leading UTF-8 byte order mark is skipped.
func hasDoctype(r io.Reader) (bool, error) {
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
	depth, roots, doctype := 0, 0, false
	for i := 0; ; i++ {
		tok, err := d.Token()
		if src.err != nil {
			return false, src.err
		}
		if errors.Is(err, io.EOF) {
			if depth != 0 || roots != 1 {
				return false, errNotXML
			}
			return doctype, nil
		}
		if err != nil {
			return false, fmt.Errorf("%w: %v", errNotXML, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots > 1 {
					return false, errNotXML
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.Trim(t, " \t\r\n")) > 0 {
				return false, errNotXML
			}
		case xml.Directive:
			if roots > 0 || doctype || !isDoctype(t) {
				return false, errNotXML
			}
			doctype = true
		case xml.ProcInst:
			if t.Target == "xml" && i > 0 {
				return false, errNotXML
			}
		}
	}
}

// isDoctype reports whether the directive body (the bytes after "<!") is a
// document type declaration: the case-sensitive keyword DOCTYPE followed
// by whitespace.
func isDoctype(d xml.Directive) bool {
	rest, ok := bytes.CutPrefix(d, []byte("DOCTYPE"))
	return ok && len(rest) > 0 && bytes.IndexByte([]byte(" \t\r\n"), rest[0]) >= 0
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
