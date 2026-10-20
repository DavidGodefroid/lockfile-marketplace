package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const content = "q /GS0 gs 0 0 10 10 re f Q BT /F1 12 Tf 72 720 Td (sample) Tj ET"

// stateDoc draws with the graphics state gs, given inline in the page
// resources.
func stateDoc(gs string) pdfxtest.Doc {
	return pdfxtest.Doc{Pages: []pdfxtest.Page{{
		Content:   content,
		Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> /ExtGState << /GS0 " + gs + " >> >>",
	}}}
}

// imageDoc draws the image object 0; Doc.Objects follow it.
func imageDoc(objects ...string) pdfxtest.Doc {
	return pdfxtest.Doc{
		Pages: []pdfxtest.Page{{
			Content:   "q 10 0 0 10 0 0 cm /Im0 Do Q",
			Resources: "<< /XObject << /Im0 {{obj 0}} >> >>",
		}},
		Objects: objects,
	}
}

func image(extra string) string {
	return pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8"+extra+" >>", []byte{0x80})
}

var mask = pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0x40})

// malformedAlpha has a /ca that is not a number.
var malformedAlpha = pdfxtest.Write(stateDoc("<< /Type /ExtGState /ca (sample) >>"))

func writeFixtures(t *testing.T) {
	write := func(name string, d pdfxtest.Doc) {
		pdfxtest.Raw(t, "testdata", name, pdfxtest.Write(d))
	}
	write("plain.pdf", stateDoc("<< /Type /ExtGState /LW 2 >>"))
	write("opaque-state.pdf", stateDoc("<< /Type /ExtGState /CA 1 /ca 1.0 /BM /Normal /SMask /None >>"))
	write("compatible-blend.pdf", stateDoc("<< /BM [/Compatible /Normal] >>"))
	write("fill-alpha.pdf", stateDoc("<< /Type /ExtGState /ca 0.5 >>"))
	write("stroke-alpha-untyped.pdf", stateDoc("<< /CA 0 >>"))
	write("blend-mode.pdf", stateDoc("<< /Type /ExtGState /BM /Multiply >>"))
	write("blend-array.pdf", stateDoc("<< /BM [/Screen /Normal] >>"))
	softMask := stateDoc("<< /Type /ExtGState /SMask << /Type /Mask /S /Luminosity /G {{obj 0}} >> >>")
	softMask.Objects = []string{pdfxtest.Stream("<< /Type /XObject /Subtype /Form /BBox [0 0 10 10] "+
		"/Group << /S /Transparency /CS /DeviceGray >> >>", []byte("0 0 10 10 re f"))}
	write("soft-mask.pdf", softMask)
	write("indirect-state.pdf", pdfxtest.Doc{
		Pages: []pdfxtest.Page{{
			Content:   content,
			Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> /ExtGState << /GS0 {{obj 0}} >> >>",
		}},
		Objects: []string{"<< /ca 0.25 >>"},
	})
	write("object-stream.pdf", pdfxtest.Doc{
		Pages: []pdfxtest.Page{{
			Content:   content,
			Resources: "<< /Font << /F1 " + pdfxtest.Helvetica + " >> /ExtGState {{obj 0}} >>",
		}},
		Objects:   []string{"<< /GS0 << /BM /Darken >> >>"},
		ObjStream: true,
	})
	write("orphan-state.pdf", pdfxtest.Doc{
		Pages:   stateDoc("<< /LW 1 >>").Pages,
		Objects: []string{"<< /Type /ExtGState /ca 0.1 >>"},
	})
	write("opaque-image.pdf", imageDoc(image("")))
	write("soft-masked-image.pdf", imageDoc(image(" /SMask {{obj 1}}"), mask))
}
