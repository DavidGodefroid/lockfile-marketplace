// Command indexgen builds the marketplace index.json from the GitHub
// releases of this repository (the JSON of `gh api .../releases`). A
// release is a plugin version when its tag is <name>/vX.Y.Z; its assets are
// those release.yml uploads: plugin.toml, one binary per platform named
// <name>-<version>-<os>-<arch>, a .sigstore.json bundle for each, and
// SHA256SUMS. A release whose title starts with "[yanked]" is listed as
// yanked. Drafts are skipped.
//
//	indexgen -releases releases.json -out index/index.json
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DavidGodefroid/lockfile/index"
	"github.com/DavidGodefroid/lockfile/pluginspec"
)

// Release is the part of a GitHub release indexgen reads.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Draft   bool    `json:"draft"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

const (
	yankedPrefix = "[yanked]"
	sumsFile     = "SHA256SUMS"
	manifestFile = "plugin.toml"
	bundleSuffix = ".sigstore.json"
	maxSums      = 1 << 20
)

// ReadReleases reads a JSON array of releases, a stream of release
// objects, or arrays of them (gh api --paginate output, slurped or not).
func ReadReleases(r io.Reader) ([]Release, error) {
	dec := json.NewDecoder(r)
	var out []Release
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); errors.Is(err, io.EOF) {
			return out, nil
		} else if err != nil {
			return nil, err
		}
		rs, err := decodeAny(raw)
		if err != nil {
			return nil, err
		}
		out = append(out, rs...)
	}
}

func decodeAny(raw json.RawMessage) ([]Release, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		var r Release
		err := json.Unmarshal(raw, &r)
		return []Release{r}, err
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	var out []Release
	for _, it := range items {
		rs, err := decodeAny(it)
		if err != nil {
			return nil, err
		}
		out = append(out, rs...)
	}
	return out, nil
}

// Build turns releases into an index. fetch downloads an asset (the
// SHA256SUMS of each release).
func Build(releases []Release, fetch func(url string) ([]byte, error), marketplace string, generated time.Time) (index.Index, error) {
	byName := map[string][]index.Version{}
	seen := map[string]bool{}
	for _, r := range releases {
		if r.Draft {
			continue
		}
		name, version, ok := strings.Cut(r.TagName, "/v")
		if !ok || !pluginspec.ValidName(name) {
			continue // not a plugin release
		}
		if !pluginspec.ValidVersion(version) {
			return index.Index{}, fmt.Errorf("%s: version %q is not X.Y.Z", r.TagName, version)
		}
		if seen[r.TagName] {
			return index.Index{}, fmt.Errorf("%s: released twice", r.TagName)
		}
		seen[r.TagName] = true
		v, err := version1(r, name, version, fetch)
		if err != nil {
			return index.Index{}, fmt.Errorf("%s: %w", r.TagName, err)
		}
		byName[name] = append(byName[name], v)
	}
	idx := index.Index{Schema: index.Schema, Marketplace: marketplace, Generated: generated.UTC(), Plugins: []index.Plugin{}}
	for name, vs := range byName {
		sort.Slice(vs, func(i, j int) bool { return less(vs[i].Version, vs[j].Version) })
		idx.Plugins = append(idx.Plugins, index.Plugin{Name: name, Versions: vs})
	}
	sort.Slice(idx.Plugins, func(i, j int) bool { return idx.Plugins[i].Name < idx.Plugins[j].Name })
	return idx, nil
}

func version1(r Release, name, version string, fetch func(string) ([]byte, error)) (index.Version, error) {
	assets := map[string]string{}
	for _, a := range r.Assets {
		assets[a.Name] = a.URL
	}
	sumsURL, ok := assets[sumsFile]
	if !ok {
		return index.Version{}, errors.New("no " + sumsFile)
	}
	b, err := fetch(sumsURL)
	if err != nil {
		return index.Version{}, err
	}
	sums, err := parseSums(b)
	if err != nil {
		return index.Version{}, err
	}
	manURL, ok := assets[manifestFile]
	if !ok || sums[manifestFile] == "" {
		return index.Version{}, errors.New("no " + manifestFile + " or no checksum for it")
	}
	v := index.Version{
		Version:        version,
		ManifestURL:    manURL,
		ManifestSHA256: sums[manifestFile],
		Yanked:         strings.HasPrefix(strings.TrimSpace(r.Name), yankedPrefix),
	}
	prefix := name + "-" + version + "-"
	linux := false
	names := make([]string, 0, len(assets))
	for n := range assets {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if !strings.HasPrefix(n, prefix) || strings.HasSuffix(n, bundleSuffix) {
			continue
		}
		os, arch, ok := strings.Cut(strings.TrimPrefix(n, prefix), "-")
		if !ok || os == "" || arch == "" || strings.Contains(arch, "-") {
			return index.Version{}, fmt.Errorf("asset %q is not <name>-<version>-<os>-<arch>", n)
		}
		bundle, ok := assets[n+bundleSuffix]
		if !ok {
			return index.Version{}, fmt.Errorf("asset %q has no signature bundle", n)
		}
		if sums[n] == "" {
			return index.Version{}, fmt.Errorf("asset %q has no checksum", n)
		}
		v.Artifacts = append(v.Artifacts, index.Artifact{OS: os, Arch: arch, URL: assets[n], SHA256: sums[n], BundleURL: bundle})
		linux = linux || os == "linux"
	}
	if !linux {
		return index.Version{}, errors.New("no linux build")
	}
	return v, nil
}

// parseSums reads sha256sum output: "<hex>  <file>" per line.
func parseSums(b []byte) (map[string]string, error) {
	out := map[string]string{}
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		sum, file, ok := strings.Cut(line, "  ")
		if !ok || len(sum) != 64 {
			return nil, fmt.Errorf("%s: malformed line %q", sumsFile, line)
		}
		out[strings.TrimPrefix(file, "*")] = sum
	}
	return out, s.Err()
}

func less(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, _ := strconv.Atoi(pa[i])
		y, _ := strconv.Atoi(pb[i])
		if x != y {
			return x < y
		}
	}
	return false
}

func httpFetch(url string) ([]byte, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("refusing non-https URL %q", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxSums+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxSums {
		return nil, fmt.Errorf("%s is larger than %d bytes", url, maxSums)
	}
	return b, nil
}

func main() {
	releases := flag.String("releases", "", "JSON of the repository's releases (gh api --paginate .../releases)")
	out := flag.String("out", "index/index.json", "index file to write")
	marketplace := flag.String("marketplace", "official", "marketplace name written in the index")
	flag.Parse()
	if *releases == "" {
		fmt.Fprintln(os.Stderr, "indexgen: -releases is required")
		os.Exit(2)
	}
	f, err := os.Open(*releases)
	if err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
	rs, err := ReadReleases(f)
	f.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
	idx, err := Build(rs, httpFetch, *marketplace, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
}
