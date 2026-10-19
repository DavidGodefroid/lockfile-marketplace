package pdfxtest

import "unicode/utf16"

// XMP wraps rdf (the content of rdf:RDF) in an XMP packet.
func XMP(rdf string) []byte {
	return []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>` +
		`<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">` +
		rdf + `</rdf:RDF></x:xmpmeta><?xpacket end="w"?>`)
}

// XMP rdf:Description samples.
const (
	PDFAIDNamespace = "http://www.aiim.org/pdfa/ns/id/"
	PDFAElement     = `<rdf:Description rdf:about="" xmlns:pdfaid="` + PDFAIDNamespace + `"><pdfaid:part>2</pdfaid:part><pdfaid:conformance>B</pdfaid:conformance></rdf:Description>`
	PDFAAttribute   = `<rdf:Description rdf:about="" xmlns:pdfaid="` + PDFAIDNamespace + `" pdfaid:part="2" pdfaid:conformance="B"/>`
	PDFAOtherPrefix = `<rdf:Description rdf:about="" xmlns:foo="` + PDFAIDNamespace + `"><foo:part>1</foo:part></rdf:Description>`
	PDFAWrongNS     = `<rdf:Description rdf:about="" xmlns:pdfaid="http://example.org/not-pdfa/"><pdfaid:part>2</pdfaid:part></rdf:Description>`
	NoPDFA          = `<rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:format>application/pdf</dc:format></rdf:Description>`
)

// MalformedXMP is an XMP packet cut in the middle of a tag.
var MalformedXMP = []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?><x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF`)

// UTF16BE encodes s as UTF-16 big endian with a byte order mark.
func UTF16BE(s []byte) []byte {
	out := []byte{0xFE, 0xFF}
	for _, u := range utf16.Encode([]rune(string(s))) {
		out = append(out, byte(u>>8), byte(u))
	}
	return out
}

// UTF16LE encodes s as UTF-16 little endian with a byte order mark.
func UTF16LE(s []byte) []byte {
	out := []byte{0xFF, 0xFE}
	for _, u := range utf16.Encode([]rune(string(s))) {
		out = append(out, byte(u), byte(u>>8))
	}
	return out
}
