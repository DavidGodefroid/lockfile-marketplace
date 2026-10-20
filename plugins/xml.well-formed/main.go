// Command xml.well-formed answers: is the file a well-formed XML document?
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
		return wellFormed(f)
	})
}

// wellFormed reports whether r holds exactly one root element, with only
// whitespace, comments and processing instructions around it, directives
// only before it and the XML declaration only as the very first token. A
// decoder error is a verdict; an error from r itself is returned. A
// leading UTF-8 byte order mark is skipped.
func wellFormed(r io.Reader) (bool, error) {
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
	depth, roots := 0, 0
	for i := 0; ; i++ {
		tok, err := d.Token()
		if src.err != nil {
			return false, src.err
		}
		if errors.Is(err, io.EOF) {
			return depth == 0 && roots == 1, nil
		}
		if err != nil {
			return false, nil
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots > 1 {
					return false, nil
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.Trim(t, " \t\r\n")) > 0 {
				return false, nil
			}
		case xml.Directive:
			if roots > 0 {
				return false, nil
			}
		case xml.ProcInst:
			if t.Target == "xml" && i > 0 {
				return false, nil
			}
		}
	}
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
