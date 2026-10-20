// Package zipx opens a plugin's input file as a ZIP archive and reads its
// entries without ever holding a whole entry in memory.
package zipx

import (
	"archive/zip"
	"errors"
	"io"
	"os"
)

// Open reads the central directory of f. An entry name archive/zip finds
// insecure (absolute, "..") is not an error here: asking about such names
// is a plugin's question.
func Open(f *os.File) (*zip.Reader, error) {
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	r, err := zip.NewReader(f, st.Size())
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return nil, err
	}
	return r, nil
}

// Prefix returns the first n bytes of the decompressed entry, fewer when
// the entry is shorter.
func Prefix(e *zip.File, n int) ([]byte, error) {
	rc, err := e.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b := make([]byte, n)
	m, err := io.ReadFull(rc, b)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return b[:m], nil
}

// Drain decompresses the whole entry and discards it; archive/zip checks
// its size and CRC-32 at the end.
func Drain(e *zip.File) error {
	rc, err := e.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(io.Discard, rc)
	return err
}
