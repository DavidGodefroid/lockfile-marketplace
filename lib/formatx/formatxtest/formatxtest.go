// Package formatxtest builds the synthetic fixtures of the non-PDF plugins'
// tests: ZIP archives, minimal OOXML packages, PNG and JPEG images with
// chosen chunks or segments, and Exif blocks. Nothing here comes from a real
// document.
package formatxtest

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Sample is a file that is in none of the formats the plugins look for
// (it is valid UTF-8 text without a BOM).
var Sample = []byte("sample text, not a structured file\n")

// Raw writes b as dir/name.
func Raw(t *testing.T, dir, name string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// Entry is one ZIP entry. Name is written as is, so it may be absolute or
// contain "..". With Raw set, Body is written as already compressed data
// with the given Method, Flags and CRC32, unchecked.
type Entry struct {
	Name   string
	Body   []byte
	Method uint16
	Flags  uint16
	Raw    bool
	CRC32  uint32
}

// File is a deflated entry.
func File(name string, body []byte) Entry {
	return Entry{Name: name, Body: body, Method: zip.Deflate}
}

// Zip returns an archive of the entries, in order.
func Zip(entries ...Entry) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.Name, Method: e.Method, Flags: e.Flags}
		if e.Raw {
			h.CRC32 = e.CRC32
			h.CompressedSize64 = uint64(len(e.Body))
			h.UncompressedSize64 = uint64(len(e.Body))
			fw, err := w.CreateRaw(h)
			if err != nil {
				panic(err)
			}
			fw.Write(e.Body)
			continue
		}
		fw, err := w.CreateHeader(h)
		if err != nil {
			panic(err)
		}
		fw.Write(e.Body)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Content types and relationship types of the minimal packages.
const (
	RelOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	RelHyperlink      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	RelAttachedTmpl   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/attachedTemplate"
	RelComments       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"

	WNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
)

// Rels returns a .rels part; each rel is {Type, Target, TargetMode}.
func Rels(rels ...[3]string) []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i, r := range rels {
		mode := ""
		if r[2] != "" {
			mode = ` TargetMode="` + r[2] + `"`
		}
		b.WriteString(`<Relationship Id="rId` + strconv.Itoa(i+1) + `" Type="` + r[0] + `" Target="` + r[1] + `"` + mode + `/>`)
	}
	b.WriteString(`</Relationships>`)
	return b.Bytes()
}

// ContentTypes returns a [Content_Types].xml with the xml and rels
// defaults and one override per {part name, content type}.
func ContentTypes(overrides ...[2]string) []byte {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	for _, o := range overrides {
		b.WriteString(`<Override PartName="` + o[0] + `" ContentType="` + o[1] + `"/>`)
	}
	b.WriteString(`</Types>`)
	return b.Bytes()
}

// DocxMain is the content type of a DOCX main part.
const DocxMain = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"

// WordXML wraps body in a w:document (or another root) in the w namespace.
func WordXML(root, body string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<w:` + root + ` xmlns:w="` + WNS + `">` + body + `</w:` + root + `>`)
}

// Paragraph is a w:body with one paragraph of sample text.
const Paragraph = `<w:body><w:p><w:r><w:t>sample</w:t></w:r></w:p></w:body>`

// Docx returns a minimal DOCX whose document body is body, plus extra
// entries.
func Docx(body string, extra ...Entry) []byte {
	entries := []Entry{
		File("[Content_Types].xml", ContentTypes([2]string{"/word/document.xml", DocxMain})),
		File("_rels/.rels", Rels([3]string{RelOfficeDocument, "word/document.xml", ""})),
		File("word/document.xml", WordXML("document", body)),
	}
	return Zip(append(entries, extra...)...)
}

// Xlsx returns a minimal XLSX package with extra content type overrides
// and entries.
func Xlsx(overrides [][2]string, extra ...Entry) []byte {
	ov := append([][2]string{{"/xl/workbook.xml", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"}}, overrides...)
	entries := []Entry{
		File("[Content_Types].xml", ContentTypes(ov...)),
		File("_rels/.rels", Rels([3]string{RelOfficeDocument, "xl/workbook.xml", ""})),
		File("xl/workbook.xml", []byte(`<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheets/></workbook>`)),
	}
	return Zip(append(entries, extra...)...)
}

func gray(w, h int) image.Image {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = uint8(i * 7)
	}
	img.Set(0, 0, color.Gray{Y: 0x80})
	return img
}

// Chunk is a PNG chunk: a 4-letter type and its data.
type Chunk struct {
	Type string
	Data []byte
}

func (c Chunk) bytes() []byte {
	b := binary.BigEndian.AppendUint32(nil, uint32(len(c.Data)))
	b = append(b, c.Type...)
	b = append(b, c.Data...)
	return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(b[4:]))
}

// PNG returns a w×h grayscale PNG with the chunks inserted after IHDR.
func PNG(w, h int, chunks ...Chunk) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, gray(w, h)); err != nil {
		panic(err)
	}
	b := buf.Bytes()
	const afterIHDR = 8 + 4 + 4 + 13 + 4
	out := append([]byte{}, b[:afterIHDR]...)
	for _, c := range chunks {
		out = append(out, c.bytes()...)
	}
	return append(out, b[afterIHDR:]...)
}

// Segment is a JPEG marker segment (0xFF, marker, length, payload).
func Segment(marker byte, payload []byte) []byte {
	b := []byte{0xFF, marker}
	b = binary.BigEndian.AppendUint16(b, uint16(len(payload)+2))
	return append(b, payload...)
}

// JPEG returns a w×h grayscale baseline JPEG with the segments inserted
// right after SOI.
func JPEG(w, h int, segments ...[]byte) []byte {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, gray(w, h), &jpeg.Options{Quality: 75}); err != nil {
		panic(err)
	}
	b := buf.Bytes()
	out := append([]byte{}, b[:2]...)
	for _, s := range segments {
		out = append(out, s...)
	}
	return append(out, b[2:]...)
}

// TIFF returns an Exif TIFF block: IFD0 holds an Orientation entry and,
// with gps, a GPS IFD pointer to a GPS IFD with a GPSVersionID only (no
// coordinates).
func TIFF(order binary.AppendByteOrder, gps bool) []byte {
	var b []byte
	if order == binary.AppendByteOrder(binary.BigEndian) {
		b = append(b, 'M', 'M')
	} else {
		b = append(b, 'I', 'I')
	}
	b = order.AppendUint16(b, 42)
	b = order.AppendUint32(b, 8)
	n := uint16(1)
	if gps {
		n = 2
	}
	b = order.AppendUint16(b, n)
	// Orientation, SHORT, 1 value: 1.
	b = order.AppendUint16(b, 0x0112)
	b = order.AppendUint16(b, 3)
	b = order.AppendUint32(b, 1)
	b = order.AppendUint16(b, 1)
	b = order.AppendUint16(b, 0)
	gpsAt := uint32(8 + 2 + int(n)*12 + 4)
	if gps {
		b = order.AppendUint16(b, 0x8825)
		b = order.AppendUint16(b, 4)
		b = order.AppendUint32(b, 1)
		b = order.AppendUint32(b, gpsAt)
	}
	b = order.AppendUint32(b, 0) // no next IFD
	if gps {
		b = order.AppendUint16(b, 1)
		b = order.AppendUint16(b, 0x0000) // GPSVersionID, BYTE, 4 values
		b = order.AppendUint16(b, 1)
		b = order.AppendUint32(b, 4)
		b = append(b, 2, 2, 0, 0)
		b = order.AppendUint32(b, 0)
	}
	return b
}

// ExifAPP1 is a JPEG APP1 Exif segment holding tiff.
func ExifAPP1(tiff []byte) []byte {
	return Segment(0xE1, append([]byte("Exif\x00\x00"), tiff...))
}
