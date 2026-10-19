// Command indexgen builds the marketplace index.json from the GitHub
// releases of this repository (the JSON of `gh api .../releases`). A
// release is a plugin version when its tag is <name>/vX.Y.Z; its assets are
// those release.yml uploads: plugin.toml and one binary per platform named
// <name>-<version>-<os>-<arch>, each with a .sigstore.json bundle.
//
// indexgen trusts no asset as such: it downloads every manifest and binary,
// verifies its bundle against the identity of release.yml on that exact tag
// (with the cosign CLI), and computes the hashes itself. A release whose
// title starts with "[yanked]" is listed as yanked. Drafts are skipped.
//
//	indexgen -releases releases.json -out index/index.json
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	manifestFile = "plugin.toml"
	bundleSuffix = ".sigstore.json"

	maxManifest = 64 << 10
	maxBundle   = 1 << 20
	maxArtifact = 64 << 20

	// identityPrefix is release.yml of this repository on a tag ref; the tag
	// itself is appended for each version.
	identityPrefix = `^https://github\.com/DavidGodefroid/lockfile-marketplace/\.github/workflows/release\.yml@refs/tags/`
	oidcIssuer     = "https://token.actions.githubusercontent.com"
)

// Fetch downloads at most max bytes from an URL.
type Fetch func(url string, max int64) ([]byte, error)

// Verify checks that bundle signs blob with a certificate whose identity
// matches the regexp.
type Verify func(blob, bundle []byte, identity string) error

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

// Build turns releases into an index.
func Build(releases []Release, fetch Fetch, verify Verify, marketplace string, generated time.Time) (index.Index, error) {
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
		v, err := indexVersion(r, name, version, fetch, verify)
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

// signed downloads an asset and its bundle, verifies the signature against
// the tag's identity and returns the asset's sha256 and the bundle URL.
func signed(assets map[string]string, file string, max int64, identity string, fetch Fetch, verify Verify) (sum, bundleURL string, err error) {
	url, ok := assets[file]
	if !ok {
		return "", "", fmt.Errorf("no asset %s", file)
	}
	bundleURL, ok = assets[file+bundleSuffix]
	if !ok {
		return "", "", fmt.Errorf("asset %s has no signature bundle", file)
	}
	blob, err := fetch(url, max)
	if err != nil {
		return "", "", err
	}
	bundle, err := fetch(bundleURL, maxBundle)
	if err != nil {
		return "", "", err
	}
	if err := verify(blob, bundle, identity); err != nil {
		return "", "", fmt.Errorf("asset %s: %w", file, err)
	}
	h := sha256.Sum256(blob)
	return hex.EncodeToString(h[:]), bundleURL, nil
}

func indexVersion(r Release, name, version string, fetch Fetch, verify Verify) (index.Version, error) {
	assets := map[string]string{}
	for _, a := range r.Assets {
		assets[a.Name] = a.URL
	}
	identity := identityPrefix + regexp.QuoteMeta(r.TagName) + "$"
	manSum, manBundle, err := signed(assets, manifestFile, maxManifest, identity, fetch, verify)
	if err != nil {
		return index.Version{}, err
	}
	v := index.Version{
		Version:           version,
		ManifestURL:       assets[manifestFile],
		ManifestSHA256:    manSum,
		ManifestBundleURL: manBundle,
		Yanked:            strings.HasPrefix(strings.TrimSpace(r.Name), yankedPrefix),
	}
	prefix := name + "-" + version + "-"
	names := make([]string, 0, len(assets))
	for n := range assets {
		names = append(names, n)
	}
	sort.Strings(names)
	linux := false
	for _, n := range names {
		if !strings.HasPrefix(n, prefix) || strings.HasSuffix(n, bundleSuffix) {
			continue
		}
		goos, goarch, ok := strings.Cut(strings.TrimPrefix(n, prefix), "-")
		if !ok || goos == "" || goarch == "" || strings.Contains(goarch, "-") {
			return index.Version{}, fmt.Errorf("asset %q is not <name>-<version>-<os>-<arch>", n)
		}
		sum, bundle, err := signed(assets, n, maxArtifact, identity, fetch, verify)
		if err != nil {
			return index.Version{}, err
		}
		v.Artifacts = append(v.Artifacts, index.Artifact{OS: goos, Arch: goarch, URL: assets[n], SHA256: sum, BundleURL: bundle})
		linux = linux || goos == "linux"
	}
	if !linux {
		return index.Version{}, errors.New("no linux build")
	}
	return v, nil
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

var httpClient = &http.Client{Timeout: 2 * time.Minute}

func httpFetch(url string, max int64) ([]byte, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("refusing non-https URL %q", url)
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("%s is larger than %d bytes", url, max)
	}
	return b, nil
}

// cosignVerify runs cosign verify-blob on temporary copies.
func cosignVerify(blob, bundle []byte, identity string) error {
	dir, err := os.MkdirTemp("", "indexgen-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	bp, kp := filepath.Join(dir, "blob"), filepath.Join(dir, "blob.sigstore.json")
	if err := os.WriteFile(bp, blob, 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(kp, bundle, 0o600); err != nil {
		return err
	}
	out, err := exec.Command("cosign", "verify-blob", "--bundle", kp,
		"--certificate-identity-regexp", identity, "--certificate-oidc-issuer", oidcIssuer, bp).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cosign verify-blob: %w: %s", err, bytes.TrimSpace(out))
	}
	return nil
}

func main() {
	releases := flag.String("releases", "", "JSON of the repository's releases (gh api --paginate .../releases)")
	out := flag.String("out", "index/index.json", "index file to write")
	marketplace := flag.String("marketplace", "official", "marketplace name written in the index")
	flag.Parse()
	if err := run(*releases, *out, *marketplace); err != nil {
		fmt.Fprintln(os.Stderr, "indexgen:", err)
		os.Exit(1)
	}
}

func run(releases, out, marketplace string) error {
	if releases == "" {
		return errors.New("-releases is required")
	}
	f, err := os.Open(releases)
	if err != nil {
		return err
	}
	rs, err := ReadReleases(f)
	f.Close()
	if err != nil {
		return err
	}
	idx, err := Build(rs, httpFetch, cosignVerify, marketplace, time.Now())
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, append(b, '\n'), 0o644)
}
