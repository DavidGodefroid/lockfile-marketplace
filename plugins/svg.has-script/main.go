// Command svg.has-script answers: does the SVG document hold a script
// element or an event-handler attribute?
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
		return hasScript(f)
	})
}

const svgNS = "http://www.w3.org/2000/svg"

// hasScript reports whether r, an SVG document, holds an element whose
// local name is script (in any namespace, so an XHTML script inside
// foreignObject counts) or an unprefixed attribute whose name starts with
// "on". The whole document is parsed first: input that is not well-formed
// XML with a single root, or whose root is not svg in the SVG namespace,
// is an error, never an answer. A leading UTF-8 byte order mark is skipped.
func hasScript(r io.Reader) (bool, error) {
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
	depth, roots, found := 0, 0, false
	for i := 0; ; i++ {
		tok, err := d.Token()
		if src.err != nil {
			return false, src.err
		}
		if errors.Is(err, io.EOF) {
			if depth != 0 || roots != 1 {
				return false, errors.New("not a well-formed XML document")
			}
			return found, nil
		}
		if err != nil {
			return false, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots > 1 {
					return false, errors.New("several root elements")
				}
				if t.Name.Space != svgNS || t.Name.Local != "svg" {
					return false, errors.New("root element is not svg in the SVG namespace")
				}
			}
			depth++
			found = found || scripted(t)
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.Trim(t, " \t\r\n")) > 0 {
				return false, errors.New("text outside the root element")
			}
		case xml.Directive:
			if roots > 0 {
				return false, errors.New("directive after the root element")
			}
		case xml.ProcInst:
			if t.Target == "xml" && i > 0 {
				return false, errors.New("misplaced XML declaration")
			}
		}
	}
}

// scripted reports whether e is a script element or carries an event
// handler attribute. Namespace declarations (xmlns:on…) and prefixed
// attributes are not event handlers.
func scripted(e xml.StartElement) bool {
	if e.Name.Local == "script" {
		return true
	}
	for _, a := range e.Attr {
		if a.Name.Space == "" && strings.HasPrefix(a.Name.Local, "on") {
			return true
		}
	}
	return false
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
