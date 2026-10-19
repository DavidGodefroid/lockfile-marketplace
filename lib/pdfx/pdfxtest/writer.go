// Package pdfxtest writes synthetic PDF files for plugin tests, byte by
// byte, so a fixture can carry exactly the structure a test needs (page
// boxes, content operators, fonts, metadata, broken tails). Content is
// neutral ("sample"); no real document is ever used.
package pdfxtest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Page sizes in points.
var (
	A4     = [4]float64{0, 0, 595.276, 841.890}
	Letter = [4]float64{0, 0, 612, 792}
	A5     = [4]float64{0, 0, 419.528, 595.276}
)

// Page is one page. Raw strings are PDF syntax and may reference
// Doc.Objects with {{obj i}}.
type Page struct {
	MediaBox, CropBox *[4]float64
	UserUnit          float64
	Rotate            *int
	Annots            []string // raw annotation dicts
	Content           string   // raw content stream; empty = no /Contents
	Resources         string   // raw resources dict; empty = << >>
}

// Doc is one PDF file.
type Doc struct {
	Pages       []Page
	MediaBox    *[4]float64 // on /Pages, inherited; nil = A4
	NoMediaBox  bool        // no MediaBox on /Pages
	PagesRotate *int
	Catalog     map[string]string // extra raw catalog entries, key with its slash
	Objects     []string          // extra objects, raw; referenced as {{obj i}}
	Metadata    []byte            // catalog XMP metadata stream

	// Layout. XRefStream writes cross-reference streams instead of classic
	// tables; ObjStream (which implies XRefStream) puts every non-stream
	// object in one object stream; Hybrid writes a classic table whose
	// trailer has /XRefStm; Linearized starts the file with a linearization
	// dictionary and a first-page section that the last startxref points to,
	// whose /Prev is the main section. The linearization is structural only
	// (no hint stream).
	XRefStream bool
	ObjStream  bool
	Hybrid     bool
	Linearized bool
}

// Raw fonts and font builders.
const (
	Helvetica = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"
	Type3     = "<< /Type /Font /Subtype /Type3 /FontBBox [0 0 1 1] /FontMatrix [0.001 0 0 0.001 0 0] " +
		"/CharProcs << >> /Encoding << /Type /Encoding /Differences [0] >> /FirstChar 0 /LastChar 0 /Widths [0] " +
		"/Resources << >> >>"
)

func descriptor(file string) string {
	d := "<< /Type /FontDescriptor /FontName /SampleSans /Flags 32 /FontBBox [0 0 1000 1000] " +
		"/ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 700 /StemV 80"
	if file != "" {
		d += " /FontFile2 " + file
	}
	return d + " >>"
}

func fontFile() string {
	return Stream("<< /Length1 19 >>", []byte("sample font program"))
}

// EmbeddedTrueType is a TrueType font with a FontFile2, as objects to put
// at index at of Doc.Objects (the font dict is Objects[at]).
func EmbeddedTrueType(at int) []string {
	return []string{
		fmt.Sprintf("<< /Type /Font /Subtype /TrueType /BaseFont /SampleSans /FirstChar 32 /LastChar 32 /Widths [250] /FontDescriptor {{obj %d}} >>", at+1),
		descriptor(fmt.Sprintf("{{obj %d}}", at+2)),
		fontFile(),
	}
}

// UnembeddedTrueType is a TrueType font whose descriptor has no font file.
func UnembeddedTrueType(at int) []string {
	return []string{
		fmt.Sprintf("<< /Type /Font /Subtype /TrueType /BaseFont /SampleSans /FirstChar 32 /LastChar 32 /Widths [250] /FontDescriptor {{obj %d}} >>", at+1),
		descriptor(""),
	}
}

// Type0 is a composite font with one CIDFontType2 descendant, embedded or
// not.
func Type0(at int, embedded bool) []string {
	objs := []string{
		fmt.Sprintf("<< /Type /Font /Subtype /Type0 /BaseFont /SampleCID /Encoding /Identity-H /DescendantFonts [{{obj %d}}] >>", at+1),
		fmt.Sprintf("<< /Type /Font /Subtype /CIDFontType2 /BaseFont /SampleCID /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor {{obj %d}} /CIDToGIDMap /Identity >>", at+2),
	}
	if embedded {
		return append(objs, descriptor(fmt.Sprintf("{{obj %d}}", at+3)), fontFile())
	}
	return append(objs, descriptor(""))
}

// Stream is a raw stream object: dict (which must not hold /Length) with
// /Length added, then data.
func Stream(dict string, data []byte) string {
	d := strings.TrimSpace(dict)
	d = strings.TrimSuffix(d, ">>")
	return fmt.Sprintf("%s /Length %d >>\nstream\n%s\nendstream", strings.TrimRight(d, " "), len(data), data)
}

func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func box(b [4]float64) string {
	return fmt.Sprintf("[%s %s %s %s]", num(b[0]), num(b[1]), num(b[2]), num(b[3]))
}

var objRef = regexp.MustCompile(`\{\{obj (\d+)\}\}`)

// Write builds the file: 1 catalog, 2 page tree, then each page and its
// content stream, then the metadata stream, then Doc.Objects; a classic
// xref table and trailer.
func Write(d Doc) []byte {
	var bodies []string
	add := func(s string) int { bodies = append(bodies, s); return len(bodies) }
	add("") // catalog, filled below
	add("") // pages
	var kids []string
	for _, p := range d.Pages {
		pageNo := add("")
		kids = append(kids, fmt.Sprintf("%d 0 R", pageNo))
		var b strings.Builder
		b.WriteString("<< /Type /Page /Parent 2 0 R")
		if p.MediaBox != nil {
			b.WriteString(" /MediaBox " + box(*p.MediaBox))
		}
		if p.CropBox != nil {
			b.WriteString(" /CropBox " + box(*p.CropBox))
		}
		if p.UserUnit != 0 {
			b.WriteString(" /UserUnit " + num(p.UserUnit))
		}
		if p.Rotate != nil {
			b.WriteString(" /Rotate " + strconv.Itoa(*p.Rotate))
		}
		res := p.Resources
		if res == "" {
			res = "<< >>"
		}
		b.WriteString(" /Resources " + res)
		if len(p.Annots) > 0 {
			b.WriteString(" /Annots [" + strings.Join(p.Annots, " ") + "]")
		}
		if p.Content != "" {
			c := add(Stream("<< >>", []byte(p.Content)))
			b.WriteString(fmt.Sprintf(" /Contents %d 0 R", c))
		}
		b.WriteString(" >>")
		bodies[pageNo-1] = b.String()
	}
	meta := 0
	if d.Metadata != nil {
		meta = add(Stream("<< /Type /Metadata /Subtype /XML >>", d.Metadata))
	}
	first := len(bodies) + 1
	for _, o := range d.Objects {
		add(o)
	}

	var cat strings.Builder
	cat.WriteString("<< /Type /Catalog /Pages 2 0 R")
	if meta != 0 {
		cat.WriteString(fmt.Sprintf(" /Metadata %d 0 R", meta))
	}
	for k, v := range d.Catalog {
		cat.WriteString(" " + k + " " + v)
	}
	cat.WriteString(" >>")
	bodies[0] = cat.String()

	var pages strings.Builder
	pages.WriteString(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d", strings.Join(kids, " "), len(d.Pages)))
	if !d.NoMediaBox {
		mb := A4
		if d.MediaBox != nil {
			mb = *d.MediaBox
		}
		pages.WriteString(" /MediaBox " + box(mb))
	}
	if d.PagesRotate != nil {
		pages.WriteString(" /Rotate " + strconv.Itoa(*d.PagesRotate))
	}
	pages.WriteString(" >>")
	bodies[1] = pages.String()

	for i, body := range bodies {
		bodies[i] = resolve(body, first)
	}
	return serialize(d, bodies, len(d.Pages))
}

func resolve(body string, first int) string {
	return objRef.ReplaceAllStringFunc(body, func(m string) string {
		n, _ := strconv.Atoi(objRef.FindStringSubmatch(m)[1])
		return fmt.Sprintf("%d 0 R", first+n)
	})
}

// entry is one cross-reference entry: type 0 free, 1 at an offset, 2 in an
// object stream (f2 = stream number, f3 = index).
type entry struct {
	typ byte
	f2  int
	f3  int
}

type out struct{ bytes.Buffer }

func (w *out) object(num int, body string) int {
	off := w.Len()
	fmt.Fprintf(w, "%d 0 obj\n%s\nendobj\n", num, body)
	return off
}

// runs returns the contiguous runs of the sorted object numbers.
func runs(nums []int) [][2]int {
	var out [][2]int
	for _, n := range nums {
		if l := len(out); l > 0 && out[l-1][0]+out[l-1][1] == n {
			out[l-1][1]++
		} else {
			out = append(out, [2]int{n, 1})
		}
	}
	return out
}

func sortedKeys(m map[int]entry) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

// section writes one cross-reference section with its trailer
// dictionary entries (trailer), as a classic table or, when streamNum > 0,
// as an xref stream object numbered streamNum (which lists itself). It
// returns the section's offset.
func (w *out) section(entries map[int]entry, size any, trailer string, streamNum int) int {
	off := w.Len()
	if streamNum > 0 {
		entries[streamNum] = entry{1, off, 0}
		keys := sortedKeys(entries)
		var data bytes.Buffer
		var index []string
		for _, r := range runs(keys) {
			index = append(index, fmt.Sprintf("%d %d", r[0], r[1]))
		}
		for _, k := range keys {
			e := entries[k]
			data.WriteByte(e.typ)
			data.Write([]byte{byte(e.f2 >> 24), byte(e.f2 >> 16), byte(e.f2 >> 8), byte(e.f2)})
			data.Write([]byte{byte(e.f3 >> 8), byte(e.f3)})
		}
		dict := fmt.Sprintf("<< /Type /XRef /Size %v /W [1 4 2] /Index [%s] %s >>", size, strings.Join(index, " "), trailer)
		w.object(streamNum, Stream(dict, data.Bytes()))
		return off
	}
	keys := sortedKeys(entries)
	w.WriteString("xref\n")
	i := 0
	for _, r := range runs(keys) {
		fmt.Fprintf(w, "%d %d\n", r[0], r[1])
		for j := 0; j < r[1]; j++ {
			e := entries[keys[i]]
			i++
			if e.typ == 0 {
				fmt.Fprintf(w, "%010d %05d f \n", e.f2, e.f3)
			} else {
				fmt.Fprintf(w, "%010d %05d n \n", e.f2, e.f3)
			}
		}
	}
	fmt.Fprintf(w, "trailer\n<< /Size %v %s >>\n", size, trailer)
	return off
}

func isStream(body string) bool { return strings.Contains(body, "\nstream\n") }

// Placeholders patched once their value is known; same width as the
// numbers they become (padded with spaces).
const (
	prevMark = "@@PREV@@"
	lenMark  = "@@LEN@@@"
	sizeMark = "@@SIZE@@"
)

func patch(b []byte, mark string, v int) {
	s := strconv.Itoa(v)
	s += strings.Repeat(" ", len(mark)-len(s))
	if i := bytes.Index(b, []byte(mark)); i >= 0 {
		copy(b[i:], s)
	}
}

func serialize(d Doc, bodies []string, pages int) []byte {
	streams := d.XRefStream || d.ObjStream
	next := len(bodies) + 1
	alloc := func() int { next++; return next - 1 }

	var w out
	w.WriteString("%PDF-1.7\n%\xe2\xe3\xcf\xd3\n")
	main := map[int]entry{0: {0, 0, 65535}}

	// Linearization dictionary and first-page section, at the start.
	firstPage := -1
	if d.Linearized {
		lin := alloc()
		off := w.object(lin, fmt.Sprintf("<< /Linearized 1 /L %s /H [0 0] /O 3 /E 0 /N %d /T 0 >>", lenMark, pages))
		fp := map[int]entry{lin: {1, off, 0}}
		fpStream := 0
		if streams {
			fpStream = alloc()
		}
		// Written before the objects exist: size and /Prev are patched.
		firstPage = w.section(fp, sizeMark, "/Root 1 0 R /Prev "+prevMark, fpStream)
	}

	var packed []int
	for i, body := range bodies {
		num := i + 1
		if d.ObjStream && !isStream(body) {
			packed = append(packed, num)
			continue
		}
		main[num] = entry{1, w.object(num, body), 0}
	}
	if len(packed) > 0 {
		stm := alloc()
		var head, data strings.Builder
		for i, num := range packed {
			fmt.Fprintf(&head, "%d %d ", num, data.Len())
			data.WriteString(bodies[num-1] + "\n")
			main[num] = entry{2, stm, i}
		}
		first := head.Len()
		content := head.String() + data.String()
		main[stm] = entry{1, w.object(stm, Stream(fmt.Sprintf("<< /Type /ObjStm /N %d /First %d >>", len(packed), first), []byte(content))), 0}
	}

	trailer := "/Root 1 0 R"
	mainStream := 0
	if streams {
		mainStream = alloc()
	}
	if d.Hybrid && !streams {
		hs := alloc()
		hEntries := map[int]entry{}
		for k, e := range main {
			if e.typ == 1 {
				hEntries[k] = e
			}
		}
		hOff := w.section(hEntries, 0, "", hs)
		main[hs] = entry{1, hOff, 0}
		trailer += fmt.Sprintf(" /XRefStm %d", hOff)
	}
	size := next
	mainOff := w.section(main, size, trailer, mainStream)
	start := mainOff
	if firstPage >= 0 {
		start = firstPage
	}
	fmt.Fprintf(&w, "startxref\n%d\n%%%%EOF\n", start)

	b := w.Bytes()
	if firstPage >= 0 {
		patch(b, prevMark, mainOff)
		patch(b, lenMark, len(b))
		patch(b, sizeMark, size)
	}
	return b
}

var (
	startXRefRe = regexp.MustCompile(`startxref\s+(\d+)`)
	rootRe      = regexp.MustCompile(`/Root (\d+ \d+ R)`)
	sizeRe      = regexp.MustCompile(`/Size (\d+)`)
)

func last(re *regexp.Regexp, b []byte) string {
	all := re.FindAllSubmatch(b, -1)
	if len(all) == 0 {
		panic("pdfxtest: no " + re.String() + " in the file")
	}
	return string(all[len(all)-1][1])
}

// AppendUpdate appends one incremental update to pdf: objects (raw,
// numbered after the file's /Size, {{obj i}} names the i-th of them), a
// cross-reference section of the same kind as the file's last one (table or
// stream) with /Prev to the previous startxref, and trailerExtra in its
// trailer.
func AppendUpdate(pdf []byte, objects []string, trailerExtra string) []byte {
	prev, _ := strconv.Atoi(last(startXRefRe, pdf))
	root := last(rootRe, pdf)
	size, _ := strconv.Atoi(last(sizeRe, pdf))
	stream := prev < len(pdf) && pdf[prev] >= '0' && pdf[prev] <= '9'

	var w out
	w.Write(pdf)
	if !bytes.HasSuffix(pdf, []byte("\n")) {
		w.WriteString("\n")
	}
	entries := map[int]entry{}
	for i, o := range objects {
		entries[size+i] = entry{1, w.object(size+i, resolve(o, size)), 0}
	}
	next := size + len(objects)
	streamNum := 0
	if stream {
		streamNum = next
		next++
	}
	trailer := fmt.Sprintf("/Root %s /Prev %d %s", root, prev, resolve(trailerExtra, size))
	off := w.section(entries, next, trailer, streamNum)
	fmt.Fprintf(&w, "startxref\n%d\n%%%%EOF\n", off)
	return w.Bytes()
}

// Sample is a file that is not a PDF.
var Sample = []byte("sample text, not a PDF\n")

const (
	textOp = "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	fonts  = "<< /Font << /F1 {{obj 0}} >> /XObject << /Im1 {{obj 1}} >> >>"
)

// pageKinds are the page contents Pages knows.
var pageKinds = map[string]string{
	"blank":           "",
	"text":            textOp,
	"white":           "1 g 0 0 595.276 841.89 re f",
	"gray":            "0.9 g 0 0 595.276 841.89 re f",
	"image":           "q 100 0 0 100 72 600 cm /Im1 Do Q",
	"invisible-text":  "BT 3 Tr /F1 12 Tf 72 720 Td (sample) Tj ET",
	"white-then-text": "1 g 0 0 595.276 841.89 re f 0 g " + textOp,
}

// Pages is an A4 document with one page per kind: blank, text, white (a
// full-page white fill), gray (a full-page light gray fill), image,
// invisible-text, white-then-text. It panics on an unknown kind.
func Pages(kinds ...string) Doc {
	d := Doc{Objects: []string{
		Helvetica,
		Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0x80}),
	}}
	for _, k := range kinds {
		c, ok := pageKinds[k]
		if !ok {
			panic("pdfxtest: unknown page kind " + k)
		}
		d.Pages = append(d.Pages, Page{Content: c, Resources: fonts})
	}
	return d
}

// Encrypt encrypts pdf with pdfcpu (AES-256). An empty user password gives
// a file anyone can open.
func Encrypt(t *testing.T, pdf []byte, user, owner string) []byte {
	t.Helper()
	conf := model.NewStatelessConfiguration()
	conf.UserPW, conf.OwnerPW = user, owner
	conf.EncryptUsingAES, conf.EncryptKeyLength = true, 256
	var out bytes.Buffer
	if err := api.Encrypt(context.Background(), bytes.NewReader(pdf), &out, conf); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return out.Bytes()
}

// Raw writes b as dir/name and returns the path.
func Raw(t *testing.T, dir, name string, b []byte) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
