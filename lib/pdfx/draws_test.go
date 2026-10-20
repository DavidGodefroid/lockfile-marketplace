package pdfx

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	drawText  = "BT /F1 12 Tf 72 720 Td (sample) Tj ET"
	drawImage = "q 10 0 0 10 72 600 cm /Im1 Do Q"
)

var drawImageObj = pdfxtest.Stream("<< /Type /XObject /Subtype /Image /Width 1 /Height 1 /ColorSpace /DeviceGray /BitsPerComponent 8 >>", []byte{0x80})

func form(res, content string) string {
	return pdfxtest.Stream("<< /Type /XObject /Subtype /Form /BBox [0 0 100 100]"+res+" >>", []byte(content))
}

// formChain is n nested forms, the deepest one drawing inner, the page
// drawing the first: objects 0 font, 1 image, 2.. the forms.
func formChain(n int, inner string) pdfxtest.Doc {
	objs := []string{pdfxtest.Helvetica, drawImageObj}
	for i := 0; i < n; i++ {
		res := " /Resources << /Font << /F1 {{obj 0}} >> /XObject << /Im1 {{obj 1}}"
		content := inner
		if i < n-1 {
			res += fmt.Sprintf(" /Fx {{obj %d}}", i+3)
			content = "/Fx Do"
		}
		objs = append(objs, form(res+" >> >>", content))
	}
	return pdfxtest.Doc{
		Objects: objs,
		Pages:   []pdfxtest.Page{{Content: "/Fx Do", Resources: "<< /XObject << /Fx {{obj 2}} >> >>"}},
	}
}

func TestPageDraws(t *testing.T) {
	res := "<< /Font << /F1 {{obj 0}} >> /XObject << /Im1 {{obj 1}} /Fm1 {{obj 2}} >> >>"
	base := []string{pdfxtest.Helvetica, drawImageObj}
	for _, c := range []struct {
		name    string
		content string
		form    string // raw object 2, if any
		want    Drawing
	}{
		{"no content", "", "", Drawing{}},
		{"graphics only", "0 g 0 0 10 10 re f", "", Drawing{}},
		{"text", drawText, "", Drawing{Text: true}},
		{"invisible text", "BT 3 Tr /F1 12 Tf (sample) Tj ET", "", Drawing{Text: true}},
		{"TJ", "BT /F1 12 Tf [(sam) 10 (ple)] TJ ET", "", Drawing{Text: true}},
		{"quote", "BT /F1 12 Tf 14 TL (sample) ' ET", "", Drawing{Text: true}},
		{"double quote", "BT /F1 12 Tf 14 TL 0 0 (sample) \" ET", "", Drawing{Text: true}},
		{"text operator in a string", "BT /F1 12 Tf (Tj) Tf ET", "", Drawing{}},
		{"image", drawImage, "", Drawing{Image: true}},
		{"escaped name", "/Im#31 Do", "", Drawing{Image: true}},
		{"inline image", "q BI /W 1 /H 1 /CS /G /BPC 8 ID \x80 EI Q", "", Drawing{Image: true}},
		{"unknown name", "/Nope Do", "", Drawing{}},
		{"text and image", drawImage + " " + drawText, "", Drawing{Text: true, Image: true}},
		{"form with text", "/Fm1 Do", form(" /Resources << /Font << /F1 {{obj 0}} >> >>", drawText), Drawing{Text: true}},
		{"form with image", "/Fm1 Do", form(" /Resources << /XObject << /Im1 {{obj 1}} >> >>", drawImage), Drawing{Image: true}},
		{"form without resources uses the page's", "/Fm1 Do", form("", drawImage), Drawing{Image: true}},
		{"form without drawing", "/Fm1 Do", form("", "0 0 10 10 re f"), Drawing{}},
		{"form not drawn", drawImage, form("", drawText), Drawing{Image: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			objs := base
			if c.form != "" {
				objs = append(append([]string{}, base...), c.form)
			}
			ctx, err := Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{
				Objects: objs,
				Pages:   []pdfxtest.Page{{Content: c.content, Resources: res}},
			})))
			if err != nil {
				t.Fatal(err)
			}
			got, err := PageDraws(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestPageDrawsDepth(t *testing.T) {
	for _, c := range []struct {
		n    int
		want bool
	}{{1, true}, {maxFormDepth, true}, {maxFormDepth + 1, false}} {
		ctx := readOnly(t, formChain(c.n, drawText))
		got, err := PageDraws(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Text != c.want {
			t.Fatalf("%d forms: text %v, want %v", c.n, got.Text, c.want)
		}
	}
}

func TestPageDrawsCycle(t *testing.T) {
	ctx := readOnly(t, pdfxtest.Doc{
		Objects: []string{
			form(" /Resources << /XObject << /Fy {{obj 1}} >> >>", "/Fy Do"),
			form(" /Resources << /XObject << /Fx {{obj 0}} >> >>", "/Fx Do /Fx Do"),
		},
		Pages: []pdfxtest.Page{{Content: "/Fx Do", Resources: "<< /XObject << /Fx {{obj 0}} >> >>"}},
	})
	got, err := PageDraws(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != (Drawing{}) {
		t.Fatalf("got %+v", got)
	}
}

// A form drawn from two pages counts on each of them.
func TestPageDrawsSharedForm(t *testing.T) {
	page := pdfxtest.Page{Content: "/Fm1 Do", Resources: "<< /XObject << /Fm1 {{obj 1}} >> >>"}
	ctx, err := Open(bytes.NewReader(pdfxtest.Write(pdfxtest.Doc{
		Objects: []string{drawImageObj, form(" /Resources << /XObject << /Im1 {{obj 0}} >> >>", drawImage)},
		Pages:   []pdfxtest.Page{page, page},
	})))
	if err != nil {
		t.Fatal(err)
	}
	for p := 1; p <= 2; p++ {
		got, err := PageDraws(ctx, p)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Image {
			t.Fatalf("page %d: got %+v", p, got)
		}
	}
}
