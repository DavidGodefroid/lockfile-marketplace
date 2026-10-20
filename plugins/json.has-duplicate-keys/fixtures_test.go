package main

import (
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	fx.Raw(t, "testdata", "top-level-dup.json", []byte("{\"a\": 1, \"b\": 2, \"a\": 3}\n"))
	fx.Raw(t, "testdata", "nested-dup.json", []byte("{\"x\": {\"y\": [1, {\"k\": null, \"k\": true}]}}"))
	fx.Raw(t, "testdata", "escaped-dup.json", []byte("{\"a\": 1, \"\\u0061\": 2}"))
	fx.Raw(t, "testdata", "dup-after-nested.json", []byte("{\"a\": {\"b\": 1}, \"a\": [2]}"))
	fx.Raw(t, "testdata", "dup-empty-key.json", []byte("{\"\": 1, \"\": 2}"))
	fx.Raw(t, "testdata", "unique.json", []byte("{\"id\": 1, \"city\": \"Liège\", \"tags\": [\"a\", {\"b\": null}], \"ok\": true}\n"))
	fx.Raw(t, "testdata", "sibling-objects.json", []byte("[{\"a\": 1}, {\"a\": 2}]"))
	fx.Raw(t, "testdata", "parent-child-same-key.json", []byte("{\"a\": {\"a\": 1}}"))
	fx.Raw(t, "testdata", "key-equals-value.json", []byte("{\"a\": \"a\", \"b\": [\"a\", \"a\"]}"))
	fx.Raw(t, "testdata", "case-differs.json", []byte("{\"a\": 1, \"A\": 2}"))
	fx.Raw(t, "testdata", "empty-object.json", []byte("{}"))
	fx.Raw(t, "testdata", "scalar.json", []byte(" 42 "))
}
