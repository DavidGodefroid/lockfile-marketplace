package main

import (
	"bytes"
	"errors"
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
	mixed := ct("multipart/mixed; boundary=x")
	harness.CheckNoAnswer(t, ".", map[string][]byte{
		"empty":               nil,
		"sample.txt":          fx.Sample,
		"sample.docx":         fx.Docx(fx.Paragraph),
		"bad content-type":    eml(ct("text/plain; =x"), "sample\r\n"),
		"bad disposition":     eml(ct("text/plain")+cd("attachment; =x"), "sample\r\n"),
		"no boundary":         eml(ct("multipart/mixed"), "sample\r\n"),
		"boundary never seen": eml(mixed, "sample\r\n"),
		"no close delimiter":  eml(mixed, "--x\r\n"+textPart),
		"attachment then cut": eml(mixed, "--x\r\n"+attachPart),
		"bad part header":     eml(mixed, "--x\r\nnot a header\r\n\r\nsample\r\n--x--\r\n"),
		"too deep":            nested(maxDepth + 1),
	})
}

func TestDepthLimit(t *testing.T) {
	if got, err := hasAttachment(bytes.NewReader(nested(maxDepth))); err != nil || got {
		t.Errorf("depth %d: hasAttachment = %v, %v; want false, nil", maxDepth, got, err)
	}
	if _, err := hasAttachment(bytes.NewReader(nested(maxDepth + 1))); !errors.Is(err, errTooDeep) {
		t.Errorf("depth %d: error = %v; want %v", maxDepth+1, err, errTooDeep)
	}
}
