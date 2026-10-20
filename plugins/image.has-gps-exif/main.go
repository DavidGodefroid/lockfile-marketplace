// Command image.has-gps-exif answers: does the PNG or JPEG image carry Exif
// GPS information?
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
		return hasGPS(f)
	})
}

var (
	pngMagic  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	exifHead  = []byte("Exif\x00\x00")
)

const (
	maxExif = 1 << 20 // cap on a PNG eXIf chunk
	gpsTag  = 0x8825  // GPSInfo IFD pointer
)

var errNotImage = errors.New("not a PNG or JPEG image")

// hasGPS reports whether r is a PNG or JPEG whose first Exif block has a
// GPSInfo tag in IFD0. The whole chunk or segment list is walked, so input in
// neither format, and any truncated or out of bounds structure, is an error.
func hasGPS(r io.Reader) (bool, error) {
	br := bufio.NewReader(r)
	head, err := br.Peek(len(pngMagic))
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	switch {
	case bytes.HasPrefix(head, pngMagic):
		_, _ = br.Discard(len(pngMagic))
		return pngGPS(br)
	case bytes.HasPrefix(head, jpegMagic):
		_, _ = br.Discard(2) // SOI; the next 0xFF starts the first segment
		return jpegGPS(br)
	}
	return false, errNotImage
}

// pngGPS walks the chunks up to IEND and answers from the first eXIf
// chunk.
func pngGPS(br *bufio.Reader) (bool, error) {
	var hdr [8]byte
	var ans *bool
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
			return ans != nil && *ans, nil
		case "eXIf":
			if n > maxExif {
				return false, fmt.Errorf("png: eXIf chunk of %d bytes exceeds %d", n, maxExif)
			}
			data := make([]byte, n)
			if _, err := io.ReadFull(br, data); err != nil {
				return false, truncated(err)
			}
			if ans == nil {
				gps, err := tiffGPS(data)
				if err != nil {
					return false, err
				}
				ans = &gps
			}
			n = 0
		}
		if err := skip(br, int64(n)+4); err != nil { // data and CRC
			return false, err
		}
	}
}

// jpegGPS walks the marker segments until SOS or EOI and answers from the
// first APP1 segment that starts with the Exif header.
func jpegGPS(br *bufio.Reader) (bool, error) {
	var ans *bool
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
			return ans != nil && *ans, nil
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
		if marker == 0xE1 && n >= len(exifHead) {
			payload := make([]byte, n)
			if _, err := io.ReadFull(br, payload); err != nil {
				return false, truncated(err)
			}
			if ans == nil && bytes.HasPrefix(payload, exifHead) {
				gps, err := tiffGPS(payload[len(exifHead):])
				if err != nil {
					return false, err
				}
				ans = &gps
			}
			continue // XMP, a later Exif or another APP1 payload
		}
		if err := skip(br, int64(n)); err != nil {
			return false, err
		}
	}
}

// tiffGPS reports whether the TIFF block b has the GPSInfo tag in IFD0.
func tiffGPS(b []byte) (bool, error) {
	if len(b) < 8 {
		return false, errors.New("tiff: header truncated")
	}
	var order binary.ByteOrder
	switch string(b[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return false, fmt.Errorf("tiff: bad byte order %q", b[:2])
	}
	if order.Uint16(b[2:4]) != 42 {
		return false, errors.New("tiff: bad magic number")
	}
	off := uint64(order.Uint32(b[4:8]))
	if off+2 > uint64(len(b)) {
		return false, fmt.Errorf("tiff: IFD0 offset %d out of bounds", off)
	}
	count := uint64(order.Uint16(b[off : off+2]))
	entries := off + 2
	if entries+count*12 > uint64(len(b)) {
		return false, fmt.Errorf("tiff: %d IFD0 entries out of bounds", count)
	}
	for i := range count {
		at := entries + i*12
		if order.Uint16(b[at:at+2]) == gpsTag {
			return true, nil
		}
	}
	return false, nil
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
