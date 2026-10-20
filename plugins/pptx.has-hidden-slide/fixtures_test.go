package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const (
	pNS       = "http://schemas.openxmlformats.org/presentationml/2006/main"
	pNSStrict = "http://purl.oclc.org/ooxml/presentationml/main"

	presMain    = "application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"
	showMacro   = "application/vnd.ms-powerpoint.slideshow.macroEnabled.main+xml"
	relSlide    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide"
	relSlideStr = "http://purl.oclc.org/ooxml/officeDocument/relationships/slide"
	relNotes    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide"
)

// slide returns a slide part with the given root attributes.
func slide(ns, attrs string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<p:sld xmlns:p="` + ns + `"` + attrs + `><p:cSld><p:spTree/></p:cSld></p:sld>`)
}

// pptx returns a presentation package: its presentation part has the given
// content type and the given relationships, plus extra entries.
func pptx(ct string, rels [][3]string, extra ...fx.Entry) []byte {
	entries := []fx.Entry{
		fx.File("[Content_Types].xml", fx.ContentTypes([2]string{"/ppt/presentation.xml", ct})),
		fx.File("_rels/.rels", fx.Rels([3]string{fx.RelOfficeDocument, "ppt/presentation.xml", ""})),
		fx.File("ppt/presentation.xml", []byte(`<p:presentation xmlns:p="`+pNS+`"><p:sldIdLst/></p:presentation>`)),
		fx.File("ppt/_rels/presentation.xml.rels", fx.Rels(rels...)),
	}
	return fx.Zip(append(entries, extra...)...)
}

// oneSlide is a presentation with ppt/slides/slide1.xml holding body.
func oneSlide(body []byte, extra ...fx.Entry) []byte {
	return pptx(presMain, [][3]string{{relSlide, "slides/slide1.xml", ""}},
		append([]fx.Entry{fx.File("ppt/slides/slide1.xml", body)}, extra...)...)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "visible.pptx", oneSlide(slide(pNS, "")))
	fx.Raw(t, "testdata", "show-true.pptx", oneSlide(slide(pNS, ` show="1"`)))
	fx.Raw(t, "testdata", "hidden.pptx", oneSlide(slide(pNS, ` show="0"`)))
	fx.Raw(t, "testdata", "hidden-false.pptx", oneSlide(slide(pNS, ` show="false"`)))
	// The strict namespace and relationship type, an absolute target.
	fx.Raw(t, "testdata", "strict-hidden.pptx", pptx(presMain,
		[][3]string{{relSlideStr, "/ppt/slides/slide1.xml", ""}},
		fx.File("ppt/slides/slide1.xml", slide(pNSStrict, ` show="0"`))))
	// A macro-enabled slide show; the second of two slides is hidden.
	fx.Raw(t, "testdata", "second-hidden.ppsm", pptx(showMacro,
		[][3]string{{relSlide, "slides/slide1.xml", ""}, {relSlide, "slides/slide2.xml", ""}},
		fx.File("ppt/slides/slide1.xml", slide(pNS, "")),
		fx.File("ppt/slides/slide2.xml", slide(pNS, ` show="0"`))))
	// A presentation without slides.
	fx.Raw(t, "testdata", "no-slides.pptx", pptx(presMain, nil))
	// show="0" on a descendant element, not the slide root.
	fx.Raw(t, "testdata", "show-elsewhere.pptx", oneSlide([]byte(
		`<p:sld xmlns:p="`+pNS+`"><p:cSld><p:spTree><p:sp show="0"/></p:spTree></p:cSld></p:sld>`)))
	// show in another namespace on the root.
	fx.Raw(t, "testdata", "foreign-show.pptx", oneSlide(slide(pNS, ` xmlns:x="urn:example" x:show="0"`)))
	// A hidden slide part no slide relationship of the presentation targets.
	fx.Raw(t, "testdata", "orphan-hidden.pptx", pptx(presMain,
		[][3]string{{relNotes, "notesSlides/notesSlide1.xml", ""}},
		fx.File("[trash]/slide9.xml", slide(pNS, ` show="0"`)),
		fx.File("ppt/notesSlides/notesSlide1.xml", slide(pNS, ` show="0"`))))
}
