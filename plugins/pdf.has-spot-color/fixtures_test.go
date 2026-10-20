package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	tint       = "<< /FunctionType 2 /Domain [0 1] /C0 [0 0 0 0] /C1 [1 0 0 0] /N 1 >>"
	separation = "[/Separation /Sample /DeviceCMYK " + tint + "]"
	deviceN    = "[/DeviceN [/Sample /Cyan] /DeviceCMYK " + tint + "]"
)

func writeFixtures(t *testing.T) {
	withCS := func(cs string) []pdfxtest.Page {
		return []pdfxtest.Page{{
			Content:   "/CS0 cs 1 scn 0 0 10 10 re f",
			Resources: "<< /ColorSpace << /CS0 " + cs + " >> >>",
		}}
	}
	write := func(name string, d pdfxtest.Doc) {
		if d.Pages == nil {
			d.Pages = []pdfxtest.Page{{Content: "0 0 10 10 re f"}}
		}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", pdfxtest.Doc{})
	write("separation.pdf", pdfxtest.Doc{Pages: withCS(separation)})
	write("devicen.pdf", pdfxtest.Doc{Pages: withCS(deviceN)})
	write("indirect.pdf", pdfxtest.Doc{Pages: withCS("{{obj 0}}"), Objects: []string{separation}})
	write("orphan.pdf", pdfxtest.Doc{Objects: []string{deviceN}})
	write("object-stream.pdf", pdfxtest.Doc{Objects: []string{separation}, ObjStream: true})
	write("cmyk-only.pdf", pdfxtest.Doc{
		Pages:   withCS("[/ICCBased {{obj 0}}]"),
		Objects: []string{pdfxtest.Stream("<< /N 4 /Alternate /DeviceCMYK >>", nil)},
	})
	// The name appears, but not as the family of a colour space array.
	write("name-not-family.pdf", pdfxtest.Doc{Objects: []string{"[/Sample /Separation /DeviceN]"}})
}
