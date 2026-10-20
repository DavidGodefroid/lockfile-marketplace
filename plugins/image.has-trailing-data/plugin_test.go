package main

import (
	"bytes"
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

func TestNoAnswer(t *testing.T) {
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":             nil,
		"sample.txt":        fx.Sample,
		"sample.docx":       fx.Docx(fx.Paragraph),
		"gif":               []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00;"),
		"png magic only":    plainPNG[:8],
		"png no IEND":       plainPNG[:len(plainPNG)-12],
		"png cut in IEND":   plainPNG[:len(plainPNG)-2],
		"png huge chunk":    cat(plainPNG[:8], []byte{0xFF, 0xFF, 0xFF, 0xFF}, []byte("IDAT")),
		"jpeg soi only":     plainJPEG[:2],
		"jpeg no EOI":       plainJPEG[:len(plainJPEG)-2],
		"jpeg cut in scan":  plainJPEG[:len(plainJPEG)-4],
		"jpeg cut in app1":  fx.JPEG(8, 8, thumbAPP1)[:40],
		"jpeg bad length":   append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x01}, make([]byte, 8)...),
		"jpeg not a marker": []byte{0xFF, 0xD8, 0xFF, 0xD8, 0x00},
		"jpeg data no ff":   []byte{0xFF, 0xD8, 0xFF, 0x12, 0x34},
	})
}

// sos is a minimal SOS segment header (one component).
var sos = fx.Segment(0xDA, []byte{1, 1, 0, 0, 63, 0})

func TestJPEGScans(t *testing.T) {
	for name, tc := range map[string]struct {
		b    []byte
		want bool
	}{
		// Stuffed 0xFF00, a restart marker and fill bytes inside a scan.
		"stuffing and restart": {cat([]byte{0xFF, 0xD8}, sos, []byte{1, 0xFF, 0x00, 2, 0xFF, 0xD3, 3, 0xFF, 0xFF, 0xD9}), false},
		// Two scans separated by a table segment, as in a progressive JPEG.
		"two scans":          {cat([]byte{0xFF, 0xD8}, sos, []byte{1, 2}, fx.Segment(0xC4, []byte{0}), sos, []byte{3, 0xFF, 0xD9}), false},
		"two scans trailing": {cat([]byte{0xFF, 0xD8}, sos, []byte{1, 2}, sos, []byte{3, 0xFF, 0xD9, 0xFF, 0xD9}), true},
		// A scan longer than the bufio buffer, without any 0xFF.
		"long scan": {cat([]byte{0xFF, 0xD8}, sos, bytes.Repeat([]byte{7}, 10000), []byte{0xFF, 0xD9, 'x'}), true},
	} {
		if got, err := hasTrailing(bytes.NewReader(tc.b)); err != nil || got != tc.want {
			t.Errorf("%s: hasTrailing = %v, %v; want %v, nil", name, got, err, tc.want)
		}
	}
	if _, err := hasTrailing(bytes.NewReader(cat([]byte{0xFF, 0xD8}, sos, []byte{1, 0xFF}))); err == nil {
		t.Error("scan ending on 0xFF: hasTrailing returned no error")
	}
}
