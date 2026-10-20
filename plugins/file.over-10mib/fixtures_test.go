package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
)

func writeFixtures(t *testing.T) {
	formatxtest.Raw(t, "testdata", "empty.bin", nil)
	formatxtest.Raw(t, "testdata", "exactly-10mib.bin", make([]byte, limit))
	formatxtest.Raw(t, "testdata", "10mib-plus-one.bin", make([]byte, limit+1))
}
