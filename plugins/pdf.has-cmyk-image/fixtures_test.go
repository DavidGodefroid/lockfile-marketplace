package main

import (
	"strconv"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

func writeFixtures(t *testing.T) {
	page := []pdfxtest.Page{{Content: "BT /F1 12 Tf 72 720 Td (sample) Tj ET", Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> >>"}}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = page
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	image := func(cs string) string {
		return pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /BitsPerComponent 8 /ColorSpace "+cs+" >>", []byte("sample"))
	}
	icc := func(n int) string {
		return pdfxtest.Stream("<< /N "+strconv.Itoa(n)+" >>", []byte("sample"))
	}
	write("plain.pdf", pdfxtest.Doc{})
	// RGB and grey images, a three-component ICC image, and DeviceCMYK
	// outside an image (a form XObject).
	write("non-cmyk.pdf", pdfxtest.Doc{Objects: []string{
		image("/DeviceRGB"),
		image("/DeviceGray"),
		image("[/ICCBased {{obj 3}}]"),
		icc(3),
		pdfxtest.Stream("<< /Type /XObject /Subtype /Form /BBox [0 0 1 1] /Group << /S /Transparency /CS /DeviceCMYK >> >>", []byte("sample")),
	}})
	write("device-cmyk.pdf", pdfxtest.Doc{Objects: []string{
		image("/DeviceCMYK"),
	}})
	write("icc-cmyk.pdf", pdfxtest.Doc{Objects: []string{
		image("[/ICCBased {{obj 1}}]"),
		icc(4),
	}})
	write("cmyk-indirect.pdf", pdfxtest.Doc{Objects: []string{
		image("{{obj 1}}"),
		"[/ICCBased {{obj 2}}]",
		icc(4),
	}, ObjStream: true})
}
