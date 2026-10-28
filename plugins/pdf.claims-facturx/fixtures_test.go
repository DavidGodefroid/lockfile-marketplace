package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const ns = "urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#"

const (
	fxElement     = `<rdf:Description rdf:about="" xmlns:fx="` + ns + `"><fx:DocumentType>INVOICE</fx:DocumentType><fx:ConformanceLevel>BASIC</fx:ConformanceLevel></rdf:Description>`
	fxAttribute   = `<rdf:Description rdf:about="" xmlns:fx="` + ns + `" fx:DocumentType="INVOICE" fx:ConformanceLevel="BASIC"/>`
	fxOtherPrefix = `<rdf:Description rdf:about="" xmlns:zf="` + ns + `"><zf:ConformanceLevel>EN 16931</zf:ConformanceLevel></rdf:Description>`
	fxWrongNS     = `<rdf:Description rdf:about="" xmlns:fx="http://example.org/not-facturx/"><fx:ConformanceLevel>BASIC</fx:ConformanceLevel></rdf:Description>`
	fxNoLevel     = `<rdf:Description rdf:about="" xmlns:fx="` + ns + `"><fx:DocumentType>INVOICE</fx:DocumentType></rdf:Description>`
)

func withXMP(x []byte) []byte {
	d := pdfxtest.Pages("text")
	d.Metadata = x
	return pdfxtest.Write(d)
}

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-metadata.pdf", pdfxtest.Write(pdfxtest.Pages("text")))
	pdfxtest.Raw(t, "testdata", "xmp-without-facturx.pdf", withXMP(pdfxtest.XMP(pdfxtest.NoPDFA)))
	pdfxtest.Raw(t, "testdata", "pdfa-only.pdf", withXMP(pdfxtest.XMP(pdfxtest.PDFAElement)))
	pdfxtest.Raw(t, "testdata", "facturx-without-level.pdf", withXMP(pdfxtest.XMP(fxNoLevel)))
	pdfxtest.Raw(t, "testdata", "facturx-element.pdf", withXMP(pdfxtest.XMP(fxElement)))
	pdfxtest.Raw(t, "testdata", "facturx-attribute.pdf", withXMP(pdfxtest.XMP(fxAttribute)))
	pdfxtest.Raw(t, "testdata", "facturx-other-prefix.pdf", withXMP(pdfxtest.XMP(fxOtherPrefix)))
	pdfxtest.Raw(t, "testdata", "facturx-wrong-namespace.pdf", withXMP(pdfxtest.XMP(fxWrongNS)))
	pdfxtest.Raw(t, "testdata", "facturx-utf16.pdf", withXMP(pdfxtest.UTF16BE(pdfxtest.XMP(fxElement))))
}
