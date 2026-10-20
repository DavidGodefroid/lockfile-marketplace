package pdfxtest

import (
	"bytes"
	"fmt"
	"strings"
)

// byteRangeMark is the /ByteRange of SignatureValue, patched by Sign; the
// numbers are written with leading zeros to keep its width.
const byteRangeMark = "[0000000000 0000000000 0000000000 0000000000]"

// SignatureField is a signature field (its widget merged in) whose value
// is objects[at+1], for Doc.Objects; put SignatureValue there. Reference
// the field ({{obj at}}) from /AcroForm /Fields.
func SignatureField(at int) string {
	return fmt.Sprintf("<< /FT /Sig /T (sample) /Type /Annot /Subtype /Widget /Rect [0 0 0 0] /V {{obj %d}} >>", at+1)
}

// SignatureValue is a signature dictionary with a zero /Contents (no real
// signature) and a /ByteRange that Sign fills in.
func SignatureValue() string {
	return "<< /Type /Sig /Filter /Adobe.PPKLite /SubFilter /adbe.pkcs7.detached /ByteRange " + byteRangeMark +
		" /Contents <" + strings.Repeat("00", 64) + "> >>"
}

// Sign sets the /ByteRange of every SignatureValue of pdf to the whole
// file except its /Contents string, as a signer does on the revision it
// signs.
func Sign(pdf []byte) []byte {
	out := bytes.Clone(pdf)
	from := 0
	for {
		i := bytes.Index(out[from:], []byte(byteRangeMark))
		if i < 0 {
			return out
		}
		i += from
		c := bytes.Index(out[i:], []byte("/Contents <"))
		if c < 0 {
			panic("pdfxtest: signature without /Contents")
		}
		start := i + c + len("/Contents ")
		end := start + bytes.IndexByte(out[start:], '>') + 1
		br := fmt.Sprintf("[%010d %010d %010d %010d]", 0, start, end, len(out)-end)
		copy(out[i:], br)
		from = end
	}
}
