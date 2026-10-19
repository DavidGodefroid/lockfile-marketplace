package pdfx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"unicode/utf16"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// PDFAIDNamespace is the XMP namespace of the PDF/A identification schema.
const PDFAIDNamespace = "http://www.aiim.org/pdfa/ns/id/"

var (
	ErrXMPTooLarge = errors.New("pdfx: XMP metadata too large")
	// ErrXMPParse carries no text of the metadata: XMP may hold personal
	// data (author, title).
	ErrXMPParse = errors.New("pdfx: XMP metadata does not parse")
	errNotXMP   = errors.New("pdfx: catalog /Metadata is not a stream")
	errOddUTF16 = errors.New("pdfx: UTF-16 text of odd length")
)

// CatalogXMP returns the decoded catalog /Metadata stream, false when there
// is none, ErrXMPTooLarge when it decodes to more than max bytes.
func CatalogXMP(ctx *model.Context, max int64) (b []byte, ok bool, err error) {
	defer catch(&err)
	cat, err := ctx.Catalog()
	if err != nil {
		return nil, false, err
	}
	o, found := cat.Find("Metadata")
	if !found || o == nil {
		return nil, false, nil
	}
	sd, _, err := ctx.DereferenceStreamDict(o)
	if err != nil {
		return nil, false, err
	}
	if sd == nil {
		return nil, false, errNotXMP
	}
	if err := sd.DecodeWithLimit(max + 1); err != nil {
		return nil, false, ErrXMPTooLarge
	}
	if int64(len(sd.Content)) > max {
		return nil, false, ErrXMPTooLarge
	}
	return sd.Content, true, nil
}

// ToUTF8 converts UTF-16 (big or little endian, with a byte order mark) to
// UTF-8 and drops a UTF-8 byte order mark; anything else passes through.
func ToUTF8(b []byte) ([]byte, error) {
	var be bool
	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		return b[3:], nil
	case bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		be = true
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}):
		be = false
	default:
		return b, nil
	}
	b = b[2:]
	if len(b)%2 != 0 {
		return nil, errOddUTF16
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		if be {
			u[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
		} else {
			u[i] = uint16(b[2*i+1])<<8 | uint16(b[2*i])
		}
	}
	return []byte(string(utf16.Decode(u))), nil
}

// HasXMPProperty reports whether the XMP has an element or an attribute
// named local in namespace ns; the namespace URI decides, not the prefix.
// It stops at the first hit and keeps no character data.
func HasXMPProperty(b []byte, ns, local string) (bool, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	// The bytes are UTF-8 already (ToUTF8); a declared encoding is ignored.
	d.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }
	for {
		t, err := d.Token()
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		if err != nil {
			return false, ErrXMPParse
		}
		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Space == ns && se.Name.Local == local {
			return true, nil
		}
		for _, a := range se.Attr {
			if a.Name.Space == ns && a.Name.Local == local {
				return true, nil
			}
		}
	}
}
