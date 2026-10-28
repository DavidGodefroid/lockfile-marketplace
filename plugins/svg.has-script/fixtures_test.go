package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

const open = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="10" height="10">`

func svg(body string) []byte {
	return []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + open + body + "</svg>\n")
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "plain.svg", svg(`<rect width="10" height="10" fill="blue"/>`))
	fx.Raw(t, "testdata", "style-and-link.svg", svg(`<style>rect{fill:red}</style><a xlink:href="#r"><rect id="r" opacity="0.5" offset="1"/></a>`))
	fx.Raw(t, "testdata", "prefixed-on-attribute.svg", svg(`<rect xlink:onclick="x" xmlns:onx="urn:sample"/>`))
	fx.Raw(t, "testdata", "script-text-only.svg", svg(`<text>script onload</text><desc><![CDATA[<script>]]></desc>`))
	fx.Raw(t, "testdata", "bom-latin1.svg", []byte("\xef\xbb\xbf<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>"+open+"<title>\xe9</title></svg>"))
	fx.Raw(t, "testdata", "script.svg", svg(`<script>var a = 1;</script>`))
	fx.Raw(t, "testdata", "nested-script.svg", svg(`<g><g><script type="text/ecmascript"/></g></g>`))
	fx.Raw(t, "testdata", "onload-root.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" onload="f()"/>`))
	fx.Raw(t, "testdata", "onclick-child.svg", svg(`<circle r="4" onclick="f()"/>`))
	fx.Raw(t, "testdata", "prefixed-root.svg", []byte(`<s:svg xmlns:s="http://www.w3.org/2000/svg"><s:script/></s:svg>`))
	fx.Raw(t, "testdata", "xhtml-script.svg", svg(`<foreignObject><body xmlns="http://www.w3.org/1999/xhtml"><script/></body></foreignObject>`))
}
