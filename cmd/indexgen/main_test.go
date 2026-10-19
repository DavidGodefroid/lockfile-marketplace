package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidGodefroid/lockfile/index"
)

const base = "https://github.com/DavidGodefroid/lockfile-marketplace/releases/download/"

func hex(c byte) string { return strings.Repeat(string(c), 64) }

// release builds a release of name@version with the assets release.yml
// uploads, and registers its SHA256SUMS in sums.
func release(name, version, title string, sums map[string]string, platforms ...string) Release {
	tag := name + "/v" + version
	dl := base + tag + "/"
	r := Release{TagName: tag, Name: title}
	var lines []string
	add := func(file, sum string) {
		r.Assets = append(r.Assets, Asset{Name: file, URL: dl + file})
		if sum != "" {
			lines = append(lines, sum+"  "+file)
		}
	}
	add("plugin.toml", hex('m'))
	add("plugin.toml.sigstore.json", hex('n'))
	for i, p := range platforms {
		file := fmt.Sprintf("%s-%s-%s", name, version, p)
		add(file, hex(byte('a'+i)))
		add(file+".sigstore.json", hex('z'))
	}
	sums[dl+"SHA256SUMS"] = strings.Join(lines, "\n") + "\n"
	add("SHA256SUMS", "")
	return r
}

func fetcher(sums map[string]string) func(string) ([]byte, error) {
	return func(url string) ([]byte, error) {
		s, ok := sums[url]
		if !ok {
			return nil, errors.New("404 " + url)
		}
		return []byte(s), nil
	}
}

var when = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func TestIndexgenFromReleases(t *testing.T) {
	sums := map[string]string{}
	rels := []Release{
		release("pdf.parses", "1.1.0", "pdf.parses 1.1.0", sums, "linux-amd64", "linux-arm64"),
		release("file.non-empty", "1.0.0", "file.non-empty 1.0.0", sums, "linux-amd64"),
		release("pdf.parses", "1.0.0", "[yanked] pdf.parses 1.0.0", sums, "linux-amd64", "linux-arm64"),
		{TagName: "pdf.parses/v2.0.0", Name: "draft", Draft: true},
		{TagName: "v0.1.0", Name: "not a plugin tag"},
	}
	idx, err := Build(rels, fetcher(sums), "official", when)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(idx)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := index.Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Marketplace != "official" || !parsed.Generated.Equal(when) || len(parsed.Plugins) != 2 {
		t.Fatalf("index %+v", parsed)
	}
	if parsed.Plugins[0].Name != "file.non-empty" || parsed.Plugins[1].Name != "pdf.parses" {
		t.Fatalf("plugins not sorted: %s, %s", parsed.Plugins[0].Name, parsed.Plugins[1].Name)
	}
	if vs := parsed.Plugins[1].Versions; len(vs) != 2 || vs[0].Version != "1.0.0" || vs[1].Version != "1.1.0" {
		t.Fatalf("versions %+v", vs)
	}
	old, err := parsed.Find("pdf.parses", "1.0.0")
	if err != nil || !old.Yanked {
		t.Fatalf("1.0.0 = %+v, %v", old, err)
	}
	v, err := parsed.Find("pdf.parses", "1.1.0")
	if err != nil || v.Yanked {
		t.Fatalf("1.1.0 = %+v, %v", v, err)
	}
	if v.ManifestURL != base+"pdf.parses/v1.1.0/plugin.toml" || v.ManifestSHA256 != hex('m') {
		t.Fatalf("manifest %q %q", v.ManifestURL, v.ManifestSHA256)
	}
	arm, ok := v.Artifact("linux", "arm64")
	if !ok || arm.SHA256 != hex('b') || arm.URL != base+"pdf.parses/v1.1.0/pdf.parses-1.1.0-linux-arm64" ||
		arm.BundleURL != arm.URL+".sigstore.json" {
		t.Fatalf("arm64 %+v", arm)
	}
	if _, err := parsed.Find("pdf.parses", "2.0.0"); err == nil {
		t.Fatal("draft indexed")
	}
}

func TestIndexgenErrors(t *testing.T) {
	cases := map[string]func(sums map[string]string) []Release{
		"no SHA256SUMS": func(sums map[string]string) []Release {
			r := release("pdf.parses", "1.0.0", "", sums, "linux-amd64")
			r.Assets = r.Assets[:len(r.Assets)-1]
			return []Release{r}
		},
		"artifact without checksum": func(sums map[string]string) []Release {
			r := release("pdf.parses", "1.0.0", "", sums, "linux-amd64")
			r.Assets = append(r.Assets, Asset{Name: "pdf.parses-1.0.0-linux-riscv64", URL: base + "x"}, Asset{Name: "pdf.parses-1.0.0-linux-riscv64.sigstore.json", URL: base + "y"})
			return []Release{r}
		},
		"artifact without bundle": func(sums map[string]string) []Release {
			r := release("pdf.parses", "1.0.0", "", sums, "linux-amd64")
			var kept []Asset
			for _, a := range r.Assets {
				if a.Name != "pdf.parses-1.0.0-linux-amd64.sigstore.json" {
					kept = append(kept, a)
				}
			}
			r.Assets = kept
			return []Release{r}
		},
		"no linux build": func(sums map[string]string) []Release {
			return []Release{release("pdf.parses", "1.0.0", "", sums, "darwin-arm64")}
		},
		"bad version": func(sums map[string]string) []Release {
			return []Release{release("pdf.parses", "1.0", "", sums, "linux-amd64")}
		},
		"duplicate": func(sums map[string]string) []Release {
			return []Release{release("pdf.parses", "1.0.0", "", sums, "linux-amd64"), release("pdf.parses", "1.0.0", "", sums, "linux-amd64")}
		},
	}
	for name, mk := range cases {
		sums := map[string]string{}
		if _, err := Build(mk(sums), fetcher(sums), "official", when); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestReadReleases(t *testing.T) {
	arr := `[{"tag_name":"a.b/v1.0.0","name":"x","draft":false,"assets":[{"name":"plugin.toml","browser_download_url":"https://e/x"}]}]`
	stream := "{\"tag_name\":\"a.b/v1.0.0\"}\n{\"tag_name\":\"a.b/v1.1.0\"}\n"
	pages := `[[{"tag_name":"a.b/v1.0.0"}],[{"tag_name":"a.b/v1.1.0"}]]`
	for in, want := range map[string]int{arr: 1, stream: 2, pages: 2} {
		rs, err := ReadReleases(strings.NewReader(in))
		if err != nil || len(rs) != want {
			t.Errorf("%q: %d releases, %v", in, len(rs), err)
		}
	}
	rs, _ := ReadReleases(strings.NewReader(arr))
	if rs[0].Assets[0].URL != "https://e/x" {
		t.Fatalf("asset %+v", rs[0].Assets[0])
	}
}
