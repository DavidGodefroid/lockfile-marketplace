package pdfx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ErrBadEncrypt is returned when the trailer's /Encrypt entry, or an
// object it references, cannot be read.
var ErrBadEncrypt = errors.New("pdfx: unreadable /Encrypt dictionary")

// Bounds of EncryptDict.
const (
	maxEncryptRefs     = 64       // indirect objects resolved per call
	maxXRefStreamBytes = 32 << 20 // decoded size of one xref stream
)

// EncryptDict returns the /Encrypt dictionary of the PDF's trailer as it is
// written in the file, or nil when the file is not encrypted. Nothing is
// decrypted (the dictionary itself never is), so it works on files that
// need a user password and on security handlers pdfcpu does not support,
// where Open fails. Indirect references inside the dictionary are resolved:
// every value of the result is direct.
//
// It walks the cross-reference chain from the last startxref (XRefChain);
// the newest trailer that has /Encrypt wins, and objects are looked up
// newest section first (classic tables and xref streams). A file without a PDF header is ErrNotPDF; a broken chain is
// one of XRefChain's errors; an /Encrypt that is not a dictionary, or that
// references a missing or malformed object, is ErrBadEncrypt.
func EncryptDict(r io.ReaderAt, size int64) (d types.Dict, err error) {
	defer catch(&err)
	ok, err := HasHeader(io.NewSectionReader(r, 0, size))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotPDF
	}
	chain, err := XRefChain(r, size)
	if err != nil {
		return nil, err
	}
	f := &rawFile{r: r, size: size}
	for _, s := range chain {
		t, err := f.trailer(s)
		if err != nil {
			return nil, err
		}
		f.sections = append(f.sections, rawSection{Section: s, trailer: t})
	}
	for _, s := range f.sections {
		enc, ok := s.trailer["Encrypt"]
		if !ok || enc == nil {
			continue
		}
		v, err := f.resolve(enc, 0, map[int]bool{})
		if err != nil {
			return nil, err
		}
		d, ok := v.(types.Dict)
		if !ok {
			return nil, ErrBadEncrypt
		}
		return d, nil
	}
	return nil, nil
}

// CryptFilterMethods returns, for each crypt filter of an /Encrypt
// dictionary's /CF, its /CFM name (None when absent): for example
// {"StdCF": "AESV2"}. A /CF that is not a dictionary of dictionaries, or a
// /CFM that is not a name, is ErrBadEncrypt. No /CF is an empty map.
func CryptFilterMethods(d types.Dict) (map[string]string, error) {
	m := map[string]string{}
	cf, ok := d["CF"]
	if !ok || cf == nil {
		return m, nil
	}
	filters, ok := cf.(types.Dict)
	if !ok {
		return nil, ErrBadEncrypt
	}
	for name, v := range filters {
		f, ok := v.(types.Dict)
		if !ok {
			return nil, ErrBadEncrypt
		}
		m[name] = "None"
		if cfm, ok := f["CFM"]; ok {
			n, ok := cfm.(types.Name)
			if !ok {
				return nil, ErrBadEncrypt
			}
			m[name] = string(n)
		}
	}
	return m, nil
}

type rawSection struct {
	Section
	trailer types.Dict // trailer or xref stream dictionary
}

type rawFile struct {
	r        io.ReaderAt
	size     int64
	sections []rawSection
	refs     int
}

// parseObject parses the object at the start of b with pdfcpu's parser.
func parseObject(b []byte) (types.Object, error) {
	s := string(b)
	return model.ParseObject(context.Background(), &s, 0, MaxDepth)
}

// trailer reads the trailer dictionary of a classic section, or the
// dictionary of an xref stream.
func (f *rawFile) trailer(s Section) (types.Dict, error) {
	var b []byte
	if s.Stream {
		head, err := readAt(f.r, f.size, s.Offset, maxDictBytes)
		if err != nil {
			return nil, err
		}
		m := objHeadRe.FindIndex(head)
		if m == nil {
			return nil, ErrBadSection
		}
		b = head[m[1]:]
	} else {
		t, err := findTrailer(f.r, f.size, s.Offset+4)
		if err != nil {
			return nil, err
		}
		if b, err = readAt(f.r, f.size, t, maxDictBytes); err != nil {
			return nil, err
		}
	}
	o, err := parseObject(b)
	if err != nil {
		return nil, ErrBadSection
	}
	d, ok := o.(types.Dict)
	if !ok {
		return nil, ErrBadSection
	}
	return d, nil
}

// resolve replaces every indirect reference in o by the object it names,
// recursively.
func (f *rawFile) resolve(o types.Object, depth int, active map[int]bool) (types.Object, error) {
	if depth > MaxDepth {
		return nil, ErrBadEncrypt
	}
	switch v := o.(type) {
	case types.IndirectRef:
		n := v.ObjectNumber.Value()
		if active[n] || f.refs >= maxEncryptRefs {
			return nil, ErrBadEncrypt
		}
		f.refs++
		obj, err := f.object(n, v.GenerationNumber.Value())
		if err != nil {
			return nil, err
		}
		active[n] = true
		defer delete(active, n)
		return f.resolve(obj, depth+1, active)
	case types.Dict:
		out := types.Dict{}
		for k, e := range v {
			r, err := f.resolve(e, depth+1, active)
			if err != nil {
				return nil, err
			}
			out[k] = r
		}
		return out, nil
	case types.Array:
		out := make(types.Array, len(v))
		for i, e := range v {
			r, err := f.resolve(e, depth+1, active)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil
	}
	return o, nil
}

// object reads object num gen, found through the newest section that has
// an in-use entry for it. Objects in object streams are not read: the
// /Encrypt dictionary may not be stored in one (PDF 32000-1 7.5.7).
func (f *rawFile) object(num, gen int) (types.Object, error) {
	for _, s := range f.sections {
		off, found, err := f.lookup(s, num)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		head, err := readAt(f.r, f.size, off, maxDictBytes)
		if err != nil {
			return nil, err
		}
		m := objHeadRe.FindSubmatchIndex(head)
		if m == nil ||
			string(head[m[2]:m[3]]) != strconv.Itoa(num) ||
			string(head[m[4]:m[5]]) != strconv.Itoa(gen) {
			return nil, ErrBadEncrypt
		}
		o, err := parseObject(head[m[1]:])
		if err != nil {
			return nil, ErrBadEncrypt
		}
		return o, nil
	}
	return nil, ErrBadEncrypt
}

// lookup returns the offset of object num in section s. An entry that is
// free or in an object stream is an error; no entry is found false.
func (f *rawFile) lookup(s rawSection, num int) (int64, bool, error) {
	if s.Stream {
		return f.lookupStream(s.Offset, s.trailer, num)
	}
	// A hybrid file's /XRefStm is not read: its classic table lists every
	// object a PDF 1.4 reader needs, the /Encrypt dictionary included.
	return f.lookupTable(s.Offset+4, num)
}

var subsectionRe = regexp.MustCompile(`^[\x00\t\n\f\r ]*(\d+)[ ]+(\d+)[ ]*\r?\n?`)
var tableEntryRe = regexp.MustCompile(`^(\d{10}) (\d{5}) ([nf])[\r\n ]{2}`)

// lookupTable scans a classic table's subsection headers from off (after
// "xref"), jumping over their 20-byte entries.
func (f *rawFile) lookupTable(off int64, num int) (int64, bool, error) {
	for range maxSections * 16 {
		head, err := readAt(f.r, f.size, off, 64)
		if err != nil {
			return 0, false, err
		}
		if bytes.HasPrefix(bytes.TrimLeft(head, "\x00\t\n\f\r "), []byte("trailer")) {
			return 0, false, nil
		}
		m := subsectionRe.FindSubmatch(head)
		if m == nil {
			return 0, false, ErrBadEncrypt
		}
		start, err1 := strconv.ParseInt(string(m[1]), 10, 64)
		count, err2 := strconv.ParseInt(string(m[2]), 10, 64)
		if err1 != nil || err2 != nil || count > f.size/20 {
			return 0, false, ErrBadEncrypt
		}
		data := off + int64(len(m[0]))
		if n := int64(num); n >= start && n < start+count {
			e, err := readAt(f.r, f.size, data+20*(n-start), 20)
			if err != nil {
				return 0, false, err
			}
			em := tableEntryRe.FindSubmatch(e)
			if em == nil || string(em[3]) != "n" {
				return 0, false, ErrBadEncrypt
			}
			o, _ := strconv.ParseInt(string(em[1]), 10, 64)
			return o, true, nil
		}
		off = data + 20*count
	}
	return 0, false, ErrBadEncrypt
}

// lookupStream decodes the xref stream whose dictionary d starts at off and
// returns the offset of object num.
func (f *rawFile) lookupStream(off int64, d types.Dict, num int) (int64, bool, error) {
	data, err := f.streamData(off, d)
	if err != nil {
		return 0, false, err
	}
	w, ok := intArray(d["W"])
	if !ok || len(w) != 3 {
		return 0, false, ErrBadEncrypt
	}
	width := 0
	for _, x := range w {
		if x < 0 || x > 8 {
			return 0, false, ErrBadEncrypt
		}
		width += x
	}
	index := []int{0, 0}
	if sz, ok := d["Size"].(types.Integer); ok {
		index[1] = sz.Value()
	}
	if v, ok := d["Index"]; ok {
		if index, ok = intArray(v); !ok || len(index)%2 != 0 {
			return 0, false, ErrBadEncrypt
		}
	}
	row := 0
	for i := 0; i < len(index); i += 2 {
		start, count := index[i], index[i+1]
		if num >= start && num < start+count {
			at := (row + num - start) * width
			if width == 0 || at+width > len(data) {
				return 0, false, ErrBadEncrypt
			}
			e := data[at : at+width]
			typ := int64(1)
			if w[0] > 0 {
				typ = field(e[:w[0]])
			}
			if typ != 1 {
				return 0, false, ErrBadEncrypt
			}
			return field(e[w[0] : w[0]+w[1]]), true, nil
		}
		row += count
	}
	return 0, false, nil
}

func field(b []byte) int64 {
	var v int64
	for _, c := range b {
		v = v<<8 | int64(c)
	}
	return v
}

func intArray(o types.Object) ([]int, bool) {
	a, ok := o.(types.Array)
	if !ok {
		return nil, false
	}
	out := make([]int, len(a))
	for i, e := range a {
		n, ok := e.(types.Integer)
		if !ok || n < 0 {
			return nil, false
		}
		out[i] = n.Value()
	}
	return out, true
}

var streamKwRe = regexp.MustCompile(`>>[\x00\t\n\f\r ]*stream(\r\n|\n|\r)`)

// streamData reads and decodes the data of the stream object at off whose
// dictionary is d. /Length must be direct.
func (f *rawFile) streamData(off int64, d types.Dict) ([]byte, error) {
	n, ok := d["Length"].(types.Integer)
	if !ok || n < 0 {
		return nil, ErrBadEncrypt
	}
	head, err := readAt(f.r, f.size, off, maxDictBytes)
	if err != nil {
		return nil, err
	}
	m := streamKwRe.FindIndex(head)
	if m == nil {
		return nil, ErrBadEncrypt
	}
	start := off + int64(m[1])
	if start+int64(n) > f.size {
		return nil, ErrBadEncrypt
	}
	raw := make([]byte, n)
	if _, err := f.r.ReadAt(raw, start); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	pipeline, err := filterPipeline(d)
	if err != nil {
		return nil, err
	}
	sd := types.NewStreamDict(d, start, nil, nil, pipeline)
	sd.Raw = raw
	if err := sd.DecodeWithLimit(maxXRefStreamBytes); err != nil {
		return nil, ErrBadEncrypt
	}
	return sd.Content, nil
}

func filterPipeline(d types.Dict) ([]types.PDFFilter, error) {
	var names types.Array
	switch v := d["Filter"].(type) {
	case nil:
		return nil, nil
	case types.Name:
		names = types.Array{v}
	case types.Array:
		names = v
	default:
		return nil, ErrBadEncrypt
	}
	var parms types.Array
	switch v := d["DecodeParms"].(type) {
	case nil:
	case types.Dict:
		parms = types.Array{v}
	case types.Array:
		parms = v
	default:
		return nil, ErrBadEncrypt
	}
	var out []types.PDFFilter
	for i, n := range names {
		name, ok := n.(types.Name)
		if !ok {
			return nil, ErrBadEncrypt
		}
		pf := types.PDFFilter{Name: string(name)}
		if i < len(parms) {
			if p, ok := parms[i].(types.Dict); ok {
				pf.DecodeParms = p
			}
		}
		out = append(out, pf)
	}
	return out, nil
}
