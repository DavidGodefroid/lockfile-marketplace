// Command rtf.has-embedded-object answers: does the RTF document embed an
// object (OLE or linked)?
package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasObject(f)
	})
}

// errNotRTF is returned when the file does not start with the RTF header.
var errNotRTF = errors.New("not an RTF file: missing {\\rtf header")

// maxWord bounds the stored control word; RTF limits words to 32 letters,
// longer runs are consumed but never match.
const maxWord = 32

// hasObject reports whether r is an RTF document holding the \object control
// word. It tokenizes control words so that an escaped backslash (\\object)
// or a longer word (\objectx) does not match, and skips the payload of
// \binN so raw binary data is never read as RTF. A document that ends early
// answers false; read errors other than io.EOF are returned.
func hasObject(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	head := make([]byte, 5)
	if _, err := io.ReadFull(br, head); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return false, errNotRTF
		}
		return false, err
	}
	if !bytes.Equal(head, []byte(`{\rtf`)) {
		return false, errNotRTF
	}
	for {
		c, err := br.ReadByte()
		if err != nil {
			return false, eof(err)
		}
		if c != '\\' {
			continue
		}
		c, err = br.ReadByte()
		if err != nil {
			return false, eof(err)
		}
		if !isLetter(c) {
			// Control symbol (\\, \{, \', ...): its one character is consumed.
			continue
		}
		word, err := readWord(br, c)
		if err != nil {
			return false, eof(err)
		}
		switch word {
		case "object":
			return true, nil
		case "bin":
			done, err := skipBin(br)
			if err != nil || done {
				return false, eof(err)
			}
		}
	}
}

// readWord reads the letters of a control word whose first letter is first
// and leaves the delimiter unread.
func readWord(br *bufio.Reader, first byte) (string, error) {
	word := []byte{first}
	for {
		c, err := br.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return string(word), nil
			}
			return "", err
		}
		if !isLetter(c) {
			return string(word), br.UnreadByte()
		}
		if len(word) <= maxWord {
			word = append(word, c)
		}
	}
}

// skipBin reads the numeric parameter of \bin and its optional space
// delimiter, then discards that many bytes. done is true when the file ends
// inside the parameter or the payload.
func skipBin(br *bufio.Reader) (done bool, err error) {
	var n int64
	for {
		c, err := br.ReadByte()
		if err != nil {
			return true, err
		}
		if c < '0' || c > '9' {
			if c != ' ' {
				if err := br.UnreadByte(); err != nil {
					return true, err
				}
			}
			break
		}
		if n > (1<<62)/10 {
			// The payload is larger than any file: everything left is data.
			_, err := io.Copy(io.Discard, br)
			return true, err
		}
		n = n*10 + int64(c-'0')
	}
	m, err := io.CopyN(io.Discard, br, n)
	if m < n {
		return true, err
	}
	return false, err
}

func isLetter(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// eof maps the end of the file to a nil error.
func eof(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
