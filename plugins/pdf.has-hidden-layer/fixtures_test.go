package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// layer is the one optional content group every fixture declares, at Objects[0].
const layer = "<< /Type /OCG /Name (sample) >>"

func doc(config string, extra ...string) []byte {
	return pdfxtest.Write(pdfxtest.Doc{
		Pages:   []pdfxtest.Page{{}},
		Catalog: map[string]string{"/OCProperties": "<< /OCGs [{{obj 0}}] /D " + config + " >>"},
		Objects: append([]string{layer}, extra...),
	})
}

// badConfig is an /OCProperties whose /D is not a dictionary.
var badConfig = doc("[{{obj 0}}]")

func writeFixtures(t *testing.T) {
	pdfxtest.Raw(t, "testdata", "no-layers.pdf", pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}}))
	pdfxtest.Raw(t, "testdata", "hidden.pdf", doc("<< /OFF [{{obj 0}}] >>"))
	pdfxtest.Raw(t, "testdata", "hidden-indirect.pdf", doc("{{obj 1}}", "<< /OFF {{obj 2}} >>", "[{{obj 0}}]"))
	pdfxtest.Raw(t, "testdata", "visible.pdf", doc("<< /ON [{{obj 0}}] >>"))
	pdfxtest.Raw(t, "testdata", "empty-off.pdf", doc("<< /ON [{{obj 0}}] /OFF [] >>"))
}
