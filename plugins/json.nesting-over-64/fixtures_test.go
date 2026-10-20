package main

import (
	"strings"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

// arrays returns n nested arrays around inner.
func arrays(n int, inner string) []byte {
	return []byte(strings.Repeat("[", n) + inner + strings.Repeat("]", n))
}

// mixed returns n alternating object/array levels, starting with an object.
func mixed(n int) []byte {
	open, closing := "", ""
	for i := range n {
		if i%2 == 0 {
			open, closing = open+`{"k":`, "}"+closing
		} else {
			open, closing = open+"[", "]"+closing
		}
	}
	return []byte(open + "1" + closing)
}

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "scalar.json", []byte("42"))
	fx.Raw(t, "testdata", "flat-object.json", []byte(`{"id": 1, "tags": ["a", {"b": null}]}`))
	fx.Raw(t, "testdata", "depth-64.json", arrays(64, "1"))
	fx.Raw(t, "testdata", "depth-65.json", arrays(65, "1"))
	fx.Raw(t, "testdata", "mixed-64.json", mixed(64))
	fx.Raw(t, "testdata", "mixed-65.json", mixed(65))
	fx.Raw(t, "testdata", "deep-then-shallow.json", []byte("["+string(arrays(65, ""))+", 1, [2]]"))
	fx.Raw(t, "testdata", "wide-shallow.json", []byte("["+strings.Repeat("[[1]],", 200)+"[]]"))
	fx.Raw(t, "testdata", "very-deep.json", arrays(100000, ""))
}
