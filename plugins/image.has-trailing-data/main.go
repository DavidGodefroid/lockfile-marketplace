// Command image.has-trailing-data answers: do bytes follow the end of the
// PNG or JPEG image?
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasTrailing(f)
	})
}

var (
	pngMagic  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
)

var errNotImage = errors.New("not a PNG or JPEG image")

// hasTrailing reports whether r is a PNG or JPEG with at least one byte
// after its end: the CRC of the IEND chunk, or the EOI marker. The end is
// found by walking the structure (chunk lengths, segment lengths and
// entropy-coded scans), never by searching for its bytes, so an IEND or EOI
// inside a chunk or segment payload (an Exif thumbnail, say) does not count.
// Input in neither format, and input that ends before IEND or EOI, is an
// error.
func hasTrailing(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	head, err := br.Peek(len(pngMagic))
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	switch {
	case bytes.HasPrefix(head, pngMagic):
		_, _ = br.Discard(len(pngMagic))
		err = pngEnd(br)
	case bytes.HasPrefix(head, jpegMagic):
		_, _ = br.Discard(2) // SOI; the next 0xFF starts the first segment
		err = jpegEnd(br)
	default:
		return false, errNotImage
	}
	if err != nil {
		return false, err
	}
	if _, err := br.ReadByte(); err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// pngEnd walks the chunks and returns once the IEND chunk, CRC included,
// has been consumed.
func pngEnd(br *bufio.Reader) error {
	var hdr [8]byte
	for {
		if _, err := io.ReadFull(br, hdr[:]); err != nil {
			return truncated("png: no IEND chunk", err)
		}
		n := binary.BigEndian.Uint32(hdr[:4])
		if n > 1<<31-1 {
			return fmt.Errorf("png: chunk length %d out of range", n)
		}
		if err := skip(br, int64(n)+4); err != nil { // data and CRC
			return truncated("png: no IEND chunk", err)
		}
		if string(hdr[4:]) == "IEND" {
			return nil
		}
	}
}

// jpegEnd walks the marker segments, and the entropy-coded data after each
// SOS, and returns once the EOI marker has been consumed.
func jpegEnd(br *bufio.Reader) error {
	var pending byte // marker already read at the end of a scan; 0 if none
	for {
		marker := pending
		pending = 0
		if marker == 0 {
			var err error
			if marker, err = readMarker(br); err != nil {
				return err
			}
		}
		switch {
		case marker == 0xD9: // EOI
			return nil
		case marker == 0x01 || marker >= 0xD0 && marker <= 0xD7: // TEM, RSTn
			continue
		case marker == 0x00 || marker == 0xD8:
			return fmt.Errorf("jpeg: unexpected marker %#02x", marker)
		}
		var l [2]byte
		if _, err := io.ReadFull(br, l[:]); err != nil {
			return truncated("jpeg: no EOI marker", err)
		}
		n := int(binary.BigEndian.Uint16(l[:]))
		if n < 2 {
			return fmt.Errorf("jpeg: segment length %d below 2", n)
		}
		if err := skip(br, int64(n-2)); err != nil {
			return truncated("jpeg: no EOI marker", err)
		}
		if marker == 0xDA { // SOS: entropy-coded data follows the header
			var err error
			if pending, err = skipScan(br); err != nil {
				return err
			}
		}
	}
}

// readMarker reads a 0xFF, any fill bytes, and returns the marker code.
func readMarker(br *bufio.Reader) (byte, error) {
	b, err := br.ReadByte()
	if err != nil {
		return 0, truncated("jpeg: no EOI marker", err)
	}
	if b != 0xFF {
		return 0, fmt.Errorf("jpeg: expected a marker, got %#02x", b)
	}
	return markerCode(br)
}

// markerCode reads past fill bytes after a 0xFF and returns the code.
func markerCode(br *bufio.Reader) (byte, error) {
	for {
		c, err := br.ReadByte()
		if err != nil {
			return 0, truncated("jpeg: no EOI marker", err)
		}
		if c != 0xFF {
			return c, nil
		}
	}
}

// skipScan consumes entropy-coded data up to the first marker that is
// neither a stuffed 0xFF00 nor a restart marker, and returns that marker.
func skipScan(br *bufio.Reader) (byte, error) {
	for {
		_, err := br.ReadSlice(0xFF)
		if errors.Is(err, bufio.ErrBufferFull) {
			continue // no 0xFF in this buffer; keep scanning
		}
		if err != nil {
			return 0, truncated("jpeg: no EOI marker", err)
		}
		c, err := markerCode(br)
		if err != nil {
			return 0, err
		}
		if c == 0x00 || c >= 0xD0 && c <= 0xD7 { // stuffed byte, RSTn
			continue
		}
		return c, nil
	}
}

// skip discards n bytes, failing if the input ends first.
func skip(br *bufio.Reader, n int64) error {
	got, err := io.CopyN(io.Discard, br, n)
	if got < n {
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}

// truncated turns an early end of input into a descriptive error.
func truncated(what string, err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return errors.New(what + ": input ends inside a structure")
	}
	return err
}
