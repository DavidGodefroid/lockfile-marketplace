// Command image.is-animated answers: is the PNG, GIF or JPEG image animated?
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image/gif"
	"image/jpeg"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return isAnimated(f)
	})
}

var (
	pngMagic   = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	gif87Magic = []byte("GIF87a")
	gif89Magic = []byte("GIF89a")
	jpegMagic  = []byte{0xFF, 0xD8, 0xFF}
)

var errNotImage = errors.New("not a PNG, GIF or JPEG image")

// isAnimated reports whether r is a GIF with more than one frame or a PNG
// with an acTL chunk. A JPEG is never animated. Input in none of the three
// formats, and any GIF or JPEG that fails to decode or PNG whose chunk list
// is broken, is an error.
func isAnimated(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	head, err := br.Peek(len(pngMagic))
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	switch {
	case bytes.HasPrefix(head, pngMagic):
		_, _ = br.Discard(len(pngMagic))
		return pngAnimated(br)
	case bytes.HasPrefix(head, gif87Magic), bytes.HasPrefix(head, gif89Magic):
		g, err := gif.DecodeAll(br)
		if err != nil {
			return false, fmt.Errorf("gif: %w", err)
		}
		return len(g.Image) > 1, nil
	case bytes.HasPrefix(head, jpegMagic):
		if _, err := jpeg.DecodeConfig(br); err != nil {
			return false, fmt.Errorf("jpeg: %w", err)
		}
		return false, nil
	}
	return false, errNotImage
}

// pngAnimated walks the chunks up to IEND and reports whether one of them is
// acTL.
func pngAnimated(br *bufio.Reader) (bool, error) {
	var hdr [8]byte
	actl := false
	for {
		if _, err := io.ReadFull(br, hdr[:]); err != nil {
			return false, truncated(err)
		}
		n := binary.BigEndian.Uint32(hdr[:4])
		if n > 1<<31-1 {
			return false, fmt.Errorf("png: chunk length %d out of range", n)
		}
		switch string(hdr[4:]) {
		case "IEND":
			return actl, nil
		case "acTL":
			actl = true
		}
		if err := skip(br, int64(n)+4); err != nil { // data and CRC
			return false, err
		}
	}
}

// skip discards n bytes, failing if the input ends first.
func skip(br *bufio.Reader, n int64) error {
	got, err := io.CopyN(io.Discard, br, n)
	if got < n {
		return truncated(err)
	}
	return nil
}

// truncated turns an early end of input into a descriptive error.
func truncated(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return errors.New("input ends inside a structure")
	}
	return err
}
