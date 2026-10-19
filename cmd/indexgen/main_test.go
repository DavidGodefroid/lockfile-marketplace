package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DavidGodefroid/lockfile/index"
)

const base = "https://github.com/DavidGodefroid/lockfile-marketplace/releases/download/"

func sum(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

// files is what the fake download serves, by URL.
type files map[string][]byte

// release builds a release of name@version with the assets release.yml
// uploads, and serves their content from fs.
func release(fs files, name, version, title string, platforms ...string) Release {
	tag := name + "/v" + version
	dl := base + tag + "/"
	r := Release{TagName: tag, Name: title}
	add := func(file string, content []byte) {
		r.Assets = append(r.Assets, Asset{Name: file, URL: dl + file})
		fs[dl+file] = content
	}
	add("plugin.toml", []byte("manifest of "+tag))
	add("plugin.toml.sigstore.json", []byte("bundle"))
	for _, p := range platforms {
		file := fmt.Sprintf("%s-%s-%s", name, version, p)
		add(file, []byte("binary "+file))
		add(file+".sigstore.json", []byte("bundle"))
	}
	add("SHA256SUMS", []byte("not read"))
	return r
}

func (fs files) fetch(url string, _ int64) ([]byte, error) {
	b, ok := fs[url]
	if !ok {
		return nil, errors.New("404 " + url)
	}
	return b, nil
}

// verifier records what it verified and fails for blobs containing a
// marker.
type verifier struct {
	fail       string
	identities map[string]string // blob → identity regexp
}

func (v *verifier) verify(blob, bundle []byte, identity string) error {
	if v.identities == nil {
		v.identities = map[string]string{}
	}
	v.identities[string(blob)] = identity
	if v.fail != "" && strings.Contains(string(blob), v.fail) {
		return errors.New("bad signature")
	}
	return nil
}

var when = time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)

func TestIndexgenFromReleases(t *testing.T) {
	fs := files{}
	rels := []Release{
		release(fs, "pdf.parses", "1.1.0", "pdf.parses 1.1.0", "linux-amd64", "linux-arm64"),
		release(fs, "file.non-empty", "1.0.0", "file.non-empty 1.0.0", "linux-amd64"),
		release(fs, "pdf.parses", "1.0.0", "[yanked] pdf.parses 1.0.0", "linux-amd64", "linux-arm64"),
		{TagName: "pdf.parses/v2.0.0", Name: "draft", Draft: true},
		{TagName: "v0.1.0", Name: "not a plugin tag"},
	}
	v := &verifier{}
	idx, err := Build(rels, fs.fetch, v.verify, "official", when)
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
	ver, err := parsed.Find("pdf.parses", "1.1.0")
	if err != nil || ver.Yanked {
		t.Fatalf("1.1.0 = %+v, %v", ver, err)
	}
	dl := base + "pdf.parses/v1.1.0/"
	if ver.ManifestURL != dl+"plugin.toml" || ver.ManifestSHA256 != sum([]byte("manifest of pdf.parses/v1.1.0")) ||
		ver.ManifestBundleURL != dl+"plugin.toml.sigstore.json" {
		t.Fatalf("manifest %+v", ver)
	}
	arm, ok := ver.Artifact("linux", "arm64")
	if !ok || arm.SHA256 != sum([]byte("binary pdf.parses-1.1.0-linux-arm64")) || arm.URL != dl+"pdf.parses-1.1.0-linux-arm64" ||
		arm.BundleURL != arm.URL+".sigstore.json" {
		t.Fatalf("arm64 %+v", arm)
	}
	if _, err := parsed.Find("pdf.parses", "2.0.0"); err == nil {
		t.Fatal("draft indexed")
	}
	// Every blob was verified against its own tag's identity.
	exact := v.identities["binary pdf.parses-1.1.0-linux-arm64"]
	re := regexp.MustCompile(exact)
	if !re.MatchString("https://github.com/DavidGodefroid/lockfile-marketplace/.github/workflows/release.yml@refs/tags/pdf.parses/v1.1.0") ||
		re.MatchString("https://github.com/DavidGodefroid/lockfile-marketplace/.github/workflows/release.yml@refs/tags/pdf.parses/v1.0.0") {
		t.Fatalf("identity %q", exact)
	}
	if len(v.identities) != 8 {
		t.Fatalf("verified %d blobs, want 8", len(v.identities))
	}
}

func TestIndexgenErrors(t *testing.T) {
	cases := map[string]func(fs files, v *verifier) []Release{
		"bad artifact signature": func(fs files, v *verifier) []Release {
			v.fail = "linux-amd64"
			return []Release{release(fs, "pdf.parses", "1.0.0", "", "linux-amd64")}
		},
		"bad manifest signature": func(fs files, v *verifier) []Release {
			v.fail = "manifest of"
			return []Release{release(fs, "pdf.parses", "1.0.0", "", "linux-amd64")}
		},
		"artifact without bundle": func(fs files, v *verifier) []Release {
			r := release(fs, "pdf.parses", "1.0.0", "", "linux-amd64")
			var kept []Asset
			for _, a := range r.Assets {
				if a.Name != "pdf.parses-1.0.0-linux-amd64.sigstore.json" {
					kept = append(kept, a)
				}
			}
			r.Assets = kept
			return []Release{r}
		},
		"manifest without bundle": func(fs files, v *verifier) []Release {
			r := release(fs, "pdf.parses", "1.0.0", "", "linux-amd64")
			var kept []Asset
			for _, a := range r.Assets {
				if a.Name != "plugin.toml.sigstore.json" {
					kept = append(kept, a)
				}
			}
			r.Assets = kept
			return []Release{r}
		},
		"no linux build": func(fs files, v *verifier) []Release {
			return []Release{release(fs, "pdf.parses", "1.0.0", "", "darwin-arm64")}
		},
		"bad version": func(fs files, v *verifier) []Release {
			return []Release{release(fs, "pdf.parses", "1.0", "", "linux-amd64")}
		},
		"duplicate": func(fs files, v *verifier) []Release {
			return []Release{release(fs, "pdf.parses", "1.0.0", "", "linux-amd64"), release(fs, "pdf.parses", "1.0.0", "", "linux-amd64")}
		},
		"download fails": func(fs files, v *verifier) []Release {
			r := release(fs, "pdf.parses", "1.0.0", "", "linux-amd64")
			delete(fs, base+"pdf.parses/v1.0.0/pdf.parses-1.0.0-linux-amd64")
			return []Release{r}
		},
	}
	for name, mk := range cases {
		fs, v := files{}, &verifier{}
		if _, err := Build(mk(fs, v), fs.fetch, v.verify, "official", when); err == nil {
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
