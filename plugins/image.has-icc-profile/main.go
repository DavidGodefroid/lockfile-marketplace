// Command image.has-icc-profile answers: does the PNG or JPEG image embed an
// ICC colour profile?
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
		return hasICC(f)
	})
}

var (
	pngMagic  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	iccHead   = []byte("ICC_PROFILE\x00")
)

var errNotImage = errors.New("not a PNG or JPEG image")

// hasICC reports whether r is a PNG with an iCCP chunk or a JPEG with an APP2
// ICC_PROFILE segment. The whole chunk or segment list is walked, so input in
// neither format, and any truncated structure, is an error.
func hasICC(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	head, err := br.Peek(len(pngMagic))
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	switch {
	case bytes.HasPrefix(head, pngMagic):
		_, _ = br.Discard(len(pngMagic))
		return pngICC(br)
	case bytes.HasPrefix(head, jpegMagic):
		_, _ = br.Discard(2) // SOI; the next 0xFF starts the first segment
		return jpegICC(br)
	}
	return false, errNotImage
}

// pngICC walks the chunks up to IEND, looking for an iCCP chunk.
func pngICC(br *bufio.Reader) (bool, error) {
	var hdr [8]byte
	found := false
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
			return found, nil
		case "iCCP":
			found = true
		}
		if err := skip(br, int64(n)+4); err != nil { // data and CRC
			return false, err
		}
	}
}

// jpegICC walks the marker segments until SOS or EOI, looking for an APP2
// segment that starts with the ICC_PROFILE identifier.
func jpegICC(br *bufio.Reader) (bool, error) {
	found := false
	for {
		b, err := br.ReadByte()
		if err != nil {
			return false, truncated(err)
		}
		if b != 0xFF {
			return false, fmt.Errorf("jpeg: expected a marker, got %#02x", b)
		}
		marker := byte(0xFF)
		for marker == 0xFF { // fill bytes
			if marker, err = br.ReadByte(); err != nil {
				return false, truncated(err)
			}
		}
		switch {
		case marker == 0xD9 || marker == 0xDA: // EOI, SOS
			return found, nil
		case marker == 0x01 || marker >= 0xD0 && marker <= 0xD7: // TEM, RSTn
			continue
		case marker == 0x00 || marker == 0xD8:
			return false, fmt.Errorf("jpeg: unexpected marker %#02x", marker)
		}
		var l [2]byte
		if _, err := io.ReadFull(br, l[:]); err != nil {
			return false, truncated(err)
		}
		n := int(binary.BigEndian.Uint16(l[:]))
		if n < 2 {
			return false, fmt.Errorf("jpeg: segment length %d below 2", n)
		}
		n -= 2
		if marker == 0xE2 && n >= len(iccHead) {
			id := make([]byte, len(iccHead))
			if _, err := io.ReadFull(br, id); err != nil {
				return false, truncated(err)
			}
			found = found || bytes.Equal(id, iccHead)
			n -= len(iccHead)
		}
		if err := skip(br, int64(n)); err != nil {
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
