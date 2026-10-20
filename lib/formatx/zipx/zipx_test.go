package zipx_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/formatxtest"
	"github.com/DavidGodefroid/lockfile-marketplace/lib/formatx/zipx"
)

func open(t *testing.T, b []byte) (*zip.Reader, error) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return zipx.Open(f)
}

func TestOpenKeepsInsecureNames(t *testing.T) {
	t.Setenv("GODEBUG", "zipinsecurepath=0")
	z, err := open(t, formatxtest.Zip(formatxtest.File("../x", formatxtest.Sample)))
	if err != nil || len(z.File) != 1 || z.File[0].Name != "../x" {
		t.Fatalf("got %v", err)
	}
}

func TestPrefixAndDrain(t *testing.T) {
	body := bytes.Repeat([]byte("sample "), 1000)
	z, err := open(t, formatxtest.Zip(formatxtest.File("a", body), formatxtest.File("b", nil)))
	if err != nil {
		t.Fatal(err)
	}
	if p, err := zipx.Prefix(z.File[0], 6); err != nil || string(p) != "sample" {
		t.Errorf("prefix: %q, %v", p, err)
	}
	if p, err := zipx.Prefix(z.File[1], 6); err != nil || len(p) != 0 {
		t.Errorf("empty prefix: %q, %v", p, err)
	}
	if err := zipx.Drain(z.File[0]); err != nil {
		t.Errorf("drain: %v", err)
	}
}

func TestDrainBadCRC(t *testing.T) {
	z, err := open(t, formatxtest.Zip(formatxtest.Entry{Name: "a", Body: []byte("sample"), Method: zip.Store, Raw: true, CRC32: 1}))
	if err != nil {
		t.Fatal(err)
	}
	if err := zipx.Drain(z.File[0]); err == nil {
		t.Error("bad CRC: no error")
	}
}

func TestNotZip(t *testing.T) {
	if _, err := open(t, formatxtest.Sample); err == nil {
		t.Error("no error")
	}
}
