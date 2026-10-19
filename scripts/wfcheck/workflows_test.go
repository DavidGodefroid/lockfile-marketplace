package wfcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const dir = "../../.github/workflows"

type step struct {
	Name string         `yaml:"name"`
	Uses string         `yaml:"uses"`
	Run  string         `yaml:"run"`
	With map[string]any `yaml:"with"`
}

type job struct {
	Needs       any               `yaml:"needs"`
	If          string            `yaml:"if"`
	Permissions map[string]string `yaml:"permissions"`
	Concurrency *struct {
		Group            string `yaml:"group"`
		CancelInProgress bool   `yaml:"cancel-in-progress"`
	} `yaml:"concurrency"`
	Steps []step `yaml:"steps"`
}

type workflow struct {
	Permissions map[string]string `yaml:"permissions"`
	Concurrency any               `yaml:"concurrency"`
	Jobs        map[string]job    `yaml:"jobs"`
}

func load(t *testing.T, name string) workflow {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	var w workflow
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w
}

func names(t *testing.T) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil || len(m) == 0 {
		t.Fatalf("no workflows in %s: %v", dir, err)
	}
	var out []string
	for _, p := range m {
		out = append(out, filepath.Base(p))
	}
	return out
}

var pinned = regexp.MustCompile(`^[\w.-]+/[\w./-]+@[0-9a-f]{40}$`)

func TestActionsPinnedBySHA(t *testing.T) {
	for _, n := range names(t) {
		for jn, j := range load(t, n).Jobs {
			for _, s := range j.Steps {
				if s.Uses != "" && !pinned.MatchString(s.Uses) {
					t.Errorf("%s/%s: %q is not pinned by commit SHA", n, jn, s.Uses)
				}
			}
		}
	}
}

func TestReadOnlyByDefault(t *testing.T) {
	for _, n := range names(t) {
		p := load(t, n).Permissions
		if len(p) != 1 || p["contents"] != "read" {
			t.Errorf("%s: top-level permissions %v, want contents: read only", n, p)
		}
	}
}

// Only the jobs that sign get an OIDC token, and only the plugin job may
// write the repository (to create the release).
func TestSigningPermissionsConfined(t *testing.T) {
	want := map[string]map[string]string{
		"release.yml/check":  nil,
		"release.yml/plugin": {"contents": "write", "id-token": "write"},
		"release.yml/index":  {"contents": "read", "id-token": "write"},
		"pages.yml/deploy":   {"actions": "read", "pages": "write", "id-token": "write"},
	}
	for _, n := range names(t) {
		for jn, j := range load(t, n).Jobs {
			key := n + "/" + jn
			w, listed := want[key]
			if !listed {
				if len(j.Permissions) != 0 {
					t.Errorf("%s: unexpected permissions %v", key, j.Permissions)
				}
				continue
			}
			if len(j.Permissions) != len(w) {
				t.Errorf("%s: permissions %v, want %v", key, j.Permissions, w)
				continue
			}
			for k, v := range w {
				if j.Permissions[k] != v {
					t.Errorf("%s: permissions %v, want %v", key, j.Permissions, w)
				}
			}
		}
	}
}

func TestTagOnMainBeforeSigning(t *testing.T) {
	w := load(t, "release.yml")
	check, ok := w.Jobs["check"]
	if !ok {
		t.Fatal("release.yml has no check job")
	}
	ancestor := false
	for _, s := range check.Steps {
		if strings.Contains(s.Run, `git merge-base --is-ancestor "$GITHUB_SHA" origin/main`) {
			ancestor = true
		}
	}
	if !ancestor {
		t.Error("check job does not verify that the tagged commit is on main")
	}
	if needs, _ := w.Jobs["plugin"].Needs.(string); needs != "check" {
		t.Errorf("plugin job needs %v, want check", w.Jobs["plugin"].Needs)
	}
	if needs, _ := w.Jobs["index"].Needs.(string); needs != "plugin" {
		t.Errorf("index job needs %v, want plugin", w.Jobs["index"].Needs)
	}
}

// A workflow- or plugin-level concurrency group would cancel a pending tag
// run; only the index job is serialised, and a pending index job is only
// ever replaced by a newer one that rebuilds from every release.
func TestConcurrencyNeverDropsATagRun(t *testing.T) {
	w := load(t, "release.yml")
	if w.Concurrency != nil {
		t.Errorf("release.yml has workflow-level concurrency %v", w.Concurrency)
	}
	for jn, j := range w.Jobs {
		if jn == "index" {
			if j.Concurrency == nil || j.Concurrency.CancelInProgress {
				t.Errorf("index job concurrency %+v, want a group with cancel-in-progress: false", j.Concurrency)
			}
			continue
		}
		if j.Concurrency != nil {
			t.Errorf("job %s has concurrency %+v", jn, j.Concurrency)
		}
	}
}

// lockfile users verify with cosign v3; cosign v2 writes the legacy bundle.
func TestCosignV3(t *testing.T) {
	seen := 0
	for _, n := range names(t) {
		for jn, j := range load(t, n).Jobs {
			for _, s := range j.Steps {
				if !strings.HasPrefix(s.Uses, "sigstore/cosign-installer@") {
					continue
				}
				seen++
				if r, _ := s.With["cosign-release"].(string); !strings.HasPrefix(r, "v3.") {
					t.Errorf("%s/%s: cosign-release %v, want v3.x", n, jn, s.With["cosign-release"])
				}
			}
		}
	}
	if seen != 3 {
		t.Errorf("%d cosign installs, want 3 (plugin, index, pages)", seen)
	}
}

func TestPagesVerifiesBeforeDeploying(t *testing.T) {
	steps := load(t, "pages.yml").Jobs["deploy"].Steps
	verified := false
	for _, s := range steps {
		if strings.Contains(s.Run, "cosign verify-blob") {
			verified = true
		}
		if strings.HasPrefix(s.Uses, "actions/upload-pages-artifact@") && !verified {
			t.Fatal("pages uploads the index before verifying its signature")
		}
	}
	if !verified {
		t.Fatal("pages never verifies the index signature")
	}
}
