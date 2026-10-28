package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// profile is a stand-in ICC profile stream: the plugin never reads its data.
var profile = pdfxtest.Stream("<< /N 3 >>", []byte("sample"))

func writeFixtures(t *testing.T) {
	write := func(name string, resources string, d pdfxtest.Doc) {
		d.Pages = []pdfxtest.Page{{Content: "/CS0 cs 0 0 0 sc 0 0 10 10 re f", Resources: resources}}
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	cs := func(space string) string { return "<< /ColorSpace << /CS0 " + space + " >> >>" }
	write("plain.pdf", "", pdfxtest.Doc{})
	write("device-rgb.pdf", cs("/DeviceRGB"), pdfxtest.Doc{})
	write("cal-rgb.pdf", cs("[/CalRGB << /WhitePoint [0.9505 1 1.089] >>]"), pdfxtest.Doc{})
	write("icc-page.pdf", cs("[/ICCBased {{obj 0}}]"), pdfxtest.Doc{Objects: []string{profile}})
	write("icc-indexed.pdf", cs("[/Indexed [/ICCBased {{obj 0}}] 1 <000000FFFFFF>]"), pdfxtest.Doc{Objects: []string{profile}})
	write("icc-indirect.pdf", cs("{{obj 1}}"), pdfxtest.Doc{Objects: []string{profile, "[/ICCBased {{obj 0}}]"}})
	write("object-stream.pdf", cs("{{obj 1}}"), pdfxtest.Doc{Objects: []string{profile, "[/ICCBased {{obj 0}}]"}, ObjStream: true})
	// An output intent carries an ICC profile but is not a colour space.
	write("output-intent-only.pdf", "", pdfxtest.Doc{
		Catalog: map[string]string{"/OutputIntents": "[<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier (sample) /DestOutputProfile {{obj 0}} >>]"},
		Objects: []string{profile},
	})
}
