package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"testing/iotest"

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
		"empty":            nil,
		"sample.txt":       fx.Sample,
		"sample.docx":      fx.Docx(fx.Paragraph),
		"truncated header": []byte(`{\rt`),
		"upper case":       []byte(`{\RTF1 \object}`),
		"leading space":    []byte(` {\rtf1 \object}`),
		"no brace":         []byte(`\rtf1 \object`),
	})
}

func TestSplitReads(t *testing.T) {
	got, err := hasObject(iotest.OneByteReader(strings.NewReader(objectRTF)))
	if err != nil || !got {
		t.Errorf("hasObject = %v, %v; want true, nil", got, err)
	}
}

func TestReadError(t *testing.T) {
	want := errors.New("boom")
	for _, prefix := range []string{`{\r`, `{\rtf1 sample`, `{\rtf1 \obj`, `{\rtf1\bin12`, `{\rtf1\bin12 ab`} {
		r := io.MultiReader(bytes.NewReader([]byte(prefix)), iotest.ErrReader(want))
		if _, err := hasObject(r); !errors.Is(err, want) {
			t.Errorf("%q: hasObject error = %v; want %v", prefix, err, want)
		}
	}
}
