package main

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
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

func TestNoAnswer(t *testing.T) {
	good := pdfxtest.Write(pdfxtest.Pages("text"))
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"sample.txt": pdfxtest.Sample,
		"empty":      nil,
		"no-eof.pdf": bytes.ReplaceAll(good, eofMarker, []byte("%%END")),
	})
}

// TestTrailingDataChunks covers a %%EOF split across two scan windows and
// junk far from the end, which the fixtures cannot place exactly.
func TestTrailingDataChunks(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
		want bool
	}{
		// "%%E" ends one window, "OF" starts the next.
		{"split marker", append(append(bytes.Repeat([]byte("x"), 10), eofMarker...), bytes.Repeat([]byte(" "), chunk-2)...), false},
		{"split marker then junk", append(append(append(bytes.Repeat([]byte("x"), 10), eofMarker...), bytes.Repeat([]byte(" "), chunk-2)...), 'x'), true},
		{"junk two chunks away", append(append(eofMarker, 'x'), bytes.Repeat([]byte("\n"), 2*chunk)...), true},
		{"marker at end", eofMarker, false},
	} {
		got, err := trailingData(bytes.NewReader(tc.data), int64(len(tc.data)))
		if err != nil || got != tc.want {
			t.Errorf("%s: got %v, %v; want %v", tc.name, got, err, tc.want)
		}
	}
	if _, err := trailingData(bytes.NewReader([]byte("sample")), 6); !errors.Is(err, errNoEOF) {
		t.Errorf("no marker: got %v, want errNoEOF", err)
	}
}
