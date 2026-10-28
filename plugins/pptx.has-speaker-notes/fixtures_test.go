package main

import (
	"strconv"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// Content types and relationship types of the PresentationML parts used
// by the fixtures.
const (
	ctPresentation = "application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"
	ctShow         = "application/vnd.openxmlformats-officedocument.presentationml.slideshow.main+xml"

	relSlide        = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide"
	relNotesSlide   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide"
	relSlideLayout  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout"
	relStrictNotes  = "http://purl.oclc.org/ooxml/officeDocument/relationships/notesSlide"
	pmlNS           = "http://schemas.openxmlformats.org/presentationml/2006/main"
	sampleSlideBody = `<p:cSld><p:spTree/></p:cSld>`
)

func pml(root string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?><p:` + root + ` xmlns:p="` + pmlNS + `">` + sampleSlideBody + `</p:` + root + `>`)
}

// pptx builds a presentation whose main part has the content type main and
// whose slides have the given relationships (nil: no rels part).
func pptx(main string, slides [][][3]string, extra ...fx.Entry) []byte {
	ov := [][2]string{{"/ppt/presentation.xml", main}}
	var presRels [][3]string
	var entries []fx.Entry
	for i, rels := range slides {
		name := "slide" + strconv.Itoa(i+1) + ".xml"
		ov = append(ov, [2]string{"/ppt/slides/" + name, slideType})
		presRels = append(presRels, [3]string{relSlide, "slides/" + name, ""})
		entries = append(entries, fx.File("ppt/slides/"+name, pml("sld")))
		if rels != nil {
			entries = append(entries, fx.File("ppt/slides/_rels/"+name+".rels", fx.Rels(rels...)))
		}
	}
	base := []fx.Entry{
		fx.File("[Content_Types].xml", fx.ContentTypes(ov...)),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "ppt/presentation.xml", ""})),
		fx.File("ppt/presentation.xml", pml("presentation")),
		fx.File("ppt/_rels/presentation.xml.rels", fx.Rels(presRels...)),
	}
	return fx.Zip(append(append(base, entries...), extra...)...)
}

var (
	layoutOnly = [][3]string{{relSlideLayout, "../slideLayouts/slideLayout1.xml", ""}}
	withNotes  = [][3]string{
		{relSlideLayout, "../slideLayouts/slideLayout1.xml", ""},
		{relNotesSlide, "../notesSlides/notesSlide1.xml", ""},
	}
	notesPart = fx.File("ppt/notesSlides/notesSlide1.xml", pml("notes"))
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "sample.pptx", pptx(ctPresentation, [][][3]string{layoutOnly}))
	// A slide with no relationships part at all.
	fx.Raw(t, "testdata", "no-slide-rels.pptx", pptx(ctPresentation, [][][3]string{nil}))
	fx.Raw(t, "testdata", "notes.pptx", pptx(ctPresentation, [][][3]string{withNotes}, notesPart))
	// Only the second slide has notes.
	fx.Raw(t, "testdata", "notes-second-slide.pptx", pptx(ctPresentation,
		[][][3]string{layoutOnly, withNotes}, notesPart))
	// A slide show (ppsx) whose slide uses the Strict relationship type.
	fx.Raw(t, "testdata", "notes-strict.ppsx", pptx(ctShow,
		[][][3]string{{{relStrictNotes, "../notesSlides/notesSlide1.xml", ""}}}, notesPart))
	// A notes master on the presentation, but no slide has notes.
	fx.Raw(t, "testdata", "notes-master-only.pptx", pptx(ctPresentation, [][][3]string{layoutOnly},
		fx.File("ppt/notesMasters/notesMaster1.xml", pml("notesMaster"))))
}
