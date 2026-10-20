package main

import (
	"os"
	"testing"

	fx "github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/sdk/harness"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

func TestMain(m *testing.M) {
	sandbox.Main()
	os.Exit(m.Run())
}

func TestConformance(t *testing.T) {
	writeFixtures(t)
	harness.Check(t, ".")
}

// TestNoAnswer checks that input that is not a JPEG, or a JPEG whose headers
// do not parse, is an error, never a 0.
func TestNoAnswer(t *testing.T) {
	cmyk := fourComponent(64, 48, adobe(0))
	bad := fourComponent(64, 48)
	bad[2+4+6+1] = 0x31 // first component sampling 3×1: unsupported
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":        nil,
		"sample.txt":   fx.Sample,
		"png":          fx.PNG(8, 8),
		"soi only":     {0xFF, 0xD8},
		"cut in sof":   cmyk[:len(cmyk)-20],
		"no sof":       {0xFF, 0xD8, 0xFF, 0xD9},
		"bad sampling": bad,
		"jpeg cut":     fx.JPEG(8, 8)[:10],
	})
}
