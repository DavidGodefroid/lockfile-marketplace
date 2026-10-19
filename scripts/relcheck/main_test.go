package main

import (
	"strings"
	"testing"
)

func manifest(name, version, question string) []byte {
	return []byte(`name = "` + name + `"
version = "` + version + `"
protocol = 1
question = "` + question + `"
one = "yes"
zero = "no"
accepts = ["application/pdf"]
limits = { cpu_s = 5, memory = "64MiB", timeout_s = 5 }
`)
}

func TestCheckVersionSameQuestion(t *testing.T) {
	tags := map[string][]byte{"1.0.0": manifest("pdf.x", "1.0.0", "Is it?")}
	if err := CheckVersion("pdf.x", "1.1.0", manifest("pdf.x", "1.1.0", "Is it?"), tags); err != nil {
		t.Fatal(err)
	}
}

func TestCheckVersionQuestionChangeNeedsMajor(t *testing.T) {
	tags := map[string][]byte{"1.0.0": manifest("pdf.x", "1.0.0", "Is it?")}
	if err := CheckVersion("pdf.x", "1.1.0", manifest("pdf.x", "1.1.0", "Is it really?"), tags); err == nil || !strings.Contains(err.Error(), "major") {
		t.Fatalf("minor bump with a changed question: %v", err)
	}
	if err := CheckVersion("pdf.x", "2.0.0", manifest("pdf.x", "2.0.0", "Is it really?"), tags); err != nil {
		t.Fatalf("major bump refused: %v", err)
	}
}

func TestCheckVersionOutOfOrder(t *testing.T) {
	// 1.1.0 released after 2.0.0 is compared with the 1.x line, not 2.0.0.
	tags := map[string][]byte{
		"1.0.0": manifest("pdf.x", "1.0.0", "Is it?"),
		"2.0.0": manifest("pdf.x", "2.0.0", "Is it really?"),
	}
	if err := CheckVersion("pdf.x", "1.1.0", manifest("pdf.x", "1.1.0", "Is it really?"), tags); err == nil {
		t.Fatal("1.x question changed to the 2.x one")
	}
	if err := CheckVersion("pdf.x", "1.1.0", manifest("pdf.x", "1.1.0", "Is it?"), tags); err != nil {
		t.Fatal(err)
	}
}

func TestCheckVersionMultilineAndQuotedKeys(t *testing.T) {
	// The comparison is on the parsed manifest, whatever the TOML spelling.
	old := manifest("pdf.x", "1.0.0", "Is it?")
	spelled := []byte(`"name" = "pdf.x"
version = "1.1.0"
protocol = 1
  question = """Is it?"""
one = 'yes'
zero = "no"
accepts = ["application/pdf"]
limits = { cpu_s = 5, memory = "64MiB", timeout_s = 5 }
`)
	if err := CheckVersion("pdf.x", "1.1.0", spelled, map[string][]byte{"1.0.0": old}); err != nil {
		t.Fatalf("same question spelled differently: %v", err)
	}
	changed := []byte(strings.Replace(string(spelled), `"""Is it?"""`, "\"\"\"Is it?\nAnd more?\"\"\"", 1))
	if err := CheckVersion("pdf.x", "1.1.0", changed, map[string][]byte{"1.0.0": old}); err == nil {
		t.Fatal("multi-line question change not seen")
	}
}

func TestCheckVersionNameAndVersion(t *testing.T) {
	if err := CheckVersion("pdf.x", "1.0.0", manifest("pdf.y", "1.0.0", "Is it?"), nil); err == nil {
		t.Fatal("manifest name differs from the tag")
	}
	if err := CheckVersion("pdf.x", "1.0.1", manifest("pdf.x", "1.0.0", "Is it?"), nil); err == nil {
		t.Fatal("manifest version differs from the tag")
	}
	if err := CheckVersion("pdf.x", "1.0", manifest("pdf.x", "1.0", "Is it?"), nil); err == nil {
		t.Fatal("bad version accepted")
	}
}

func TestCheckVersionReleasedIsFrozen(t *testing.T) {
	tags := map[string][]byte{"1.0.0": manifest("pdf.x", "1.0.0", "Is it?")}
	if err := CheckVersion("pdf.x", "1.0.0", manifest("pdf.x", "1.0.0", "Is it?"), tags); err != nil {
		t.Fatal(err)
	}
	other := []byte(strings.Replace(string(manifest("pdf.x", "1.0.0", "Is it?")), "timeout_s = 5", "timeout_s = 6", 1))
	if err := CheckVersion("pdf.x", "1.0.0", other, tags); err == nil {
		t.Fatal("released manifest changed without a new version")
	}
}

func TestParseTag(t *testing.T) {
	name, version, err := ParseTag("pdf.has-blank-page/v1.2.3")
	if err != nil || name != "pdf.has-blank-page" || version != "1.2.3" {
		t.Fatalf("%q %q %v", name, version, err)
	}
	for _, bad := range []string{"v1.0.0", "pdf.x/1.0.0", "pdf.x/v1.0", "../x/v1.0.0", "a/b/v1.0.0"} {
		if _, _, err := ParseTag(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
