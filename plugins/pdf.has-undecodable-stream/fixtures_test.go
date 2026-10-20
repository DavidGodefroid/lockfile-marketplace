package main

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func deflate(t testing.TB, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zlib.NewWriter(&out)
	if _, err := w.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

// pdf is a one-page file with extra objects, in the given layout.
func pdf(objects []string, objStream bool) []byte {
	return pdfxtest.Write(pdfxtest.Doc{
		Pages:     []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}},
		Objects:   objects,
		ObjStream: objStream,
	})
}

func writeFixtures(t *testing.T) {
	sample := []byte("sample")
	flate := deflate(t, sample)
	corrupt := pdfxtest.Stream("<< /Filter /FlateDecode >>", sample)
	write := func(name string, objects []string, objStream bool) {
		pdfxtest.Raw(t, "testdata", name, pdf(objects, objStream))
	}
	write("plain.pdf", nil, false)
	write("flate.pdf", []string{pdfxtest.Stream("<< /Filter /FlateDecode >>", flate)}, false)
	write("filter-chain.pdf", []string{pdfxtest.Stream("<< /Filter [/ASCIIHexDecode /FlateDecode] >>",
		[]byte(hex.EncodeToString(flate)+">"))}, false)
	// The object stream and the cross-reference stream decode.
	write("object-stream.pdf", []string{"<< /Type /Sample >>"}, true)
	// Image codecs are not decoded.
	write("dct-image.pdf", []string{pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 /Filter /DCTDecode >>", sample)}, false)
	write("ccitt-image.pdf", []string{pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 1 /Filter [/FlateDecode /CCITTFaxDecode] >>", sample)}, false)

	write("corrupt-flate.pdf", []string{corrupt}, false)
	write("corrupt-in-object-stream-file.pdf", []string{"<< /Type /Sample >>", corrupt}, true)
	write("corrupt-hex.pdf", []string{pdfxtest.Stream("<< /Filter /ASCIIHexDecode >>", []byte("zz>"))}, false)
	write("corrupt-chain.pdf", []string{pdfxtest.Stream("<< /Filter [/ASCIIHexDecode /FlateDecode] >>",
		[]byte(hex.EncodeToString(sample)+">"))}, false)
	write("unknown-filter.pdf", []string{pdfxtest.Stream("<< /Filter /SampleDecode >>", sample)}, false)
}
