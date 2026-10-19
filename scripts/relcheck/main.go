// Command relcheck enforces the versioning rules of plugin manifests,
// against the plugin's release tags (<name>/vX.Y.Z) in git:
//   - the manifest's name is the plugin's directory and tag name, its
//     version the tag's;
//   - every version of the same major asks the same question with the same
//     meaning of 1 and 0 (a changed question is a new major version);
//   - a released version's manifest never changes.
//
// The manifests are compared parsed, whatever their TOML spelling.
//
//	go run ./scripts/relcheck tag <name>/vX.Y.Z   (release.yml)
//	go run ./scripts/relcheck all                 (ci.yml)
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/DavidGodefroid/lockfile/pluginspec"
)

// ParseTag splits <name>/vX.Y.Z.
func ParseTag(tag string) (name, version string, err error) {
	name, version, ok := strings.Cut(tag, "/v")
	if !ok || !pluginspec.ValidName(name) || !pluginspec.ValidVersion(version) {
		return "", "", fmt.Errorf("tag %q is not <plugin>/vX.Y.Z", tag)
	}
	return name, version, nil
}

func major(v string) string { return v[:strings.IndexByte(v, '.')] }

// CheckVersion checks the manifest of name@version against the manifests of
// the plugin's released versions (version → plugin.toml).
func CheckVersion(name, version string, manifest []byte, released map[string][]byte) error {
	if !pluginspec.ValidVersion(version) {
		return fmt.Errorf("%s: version %q is not X.Y.Z", name, version)
	}
	m, err := pluginspec.ParseManifest(manifest)
	if err != nil {
		return fmt.Errorf("%s@%s: %w", name, version, err)
	}
	if m.Name != name || m.Version != version {
		return fmt.Errorf("%s@%s: the manifest says %s@%s", name, version, m.Name, m.Version)
	}
	var errs []error
	for v, b := range released {
		old, err := pluginspec.ParseManifest(b)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s@%s (released): %w", name, v, err))
			continue
		}
		if v == version && !reflect.DeepEqual(old, m) {
			errs = append(errs, fmt.Errorf("%s@%s is released: its manifest cannot change, release a new version", name, version))
		}
		if major(v) == major(version) && (old.Question != m.Question || old.One != m.One || old.Zero != m.Zero) {
			errs = append(errs, fmt.Errorf("%s@%s: question, one or zero differ from %s: a changed question needs a new major version", name, version, v))
		}
	}
	return errors.Join(errs...)
}

func git(args ...string) ([]byte, error) {
	var out, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out.Bytes(), nil
}

// released reads the manifest of every release tag of a plugin.
func released(name string) (map[string][]byte, error) {
	out, err := git("tag", "-l", name+"/v*")
	if err != nil {
		return nil, err
	}
	tags := map[string][]byte{}
	for _, tag := range strings.Fields(string(out)) {
		_, v, err := ParseTag(tag)
		if err != nil {
			continue
		}
		b, err := git("show", tag+":plugins/"+name+"/plugin.toml")
		if err != nil {
			return nil, err
		}
		tags[v] = b
	}
	return tags, nil
}

func checkPlugin(name, version string) error {
	b, err := os.ReadFile(filepath.Join("plugins", name, "plugin.toml"))
	if err != nil {
		return err
	}
	if version == "" {
		m, err := pluginspec.ParseManifest(b)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		version = m.Version
	}
	tags, err := released(name)
	if err != nil {
		return err
	}
	return CheckVersion(name, version, b, tags)
}

func run(args []string) error {
	switch {
	case len(args) == 2 && args[0] == "tag":
		name, version, err := ParseTag(args[1])
		if err != nil {
			return err
		}
		return checkPlugin(name, version)
	case len(args) == 1 && args[0] == "all":
		dirs, err := os.ReadDir("plugins")
		if err != nil {
			return err
		}
		var errs []error
		for _, d := range dirs {
			if d.IsDir() {
				errs = append(errs, checkPlugin(d.Name(), ""))
			}
		}
		return errors.Join(errs...)
	}
	return errors.New("usage: relcheck tag <name>/vX.Y.Z | relcheck all")
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "relcheck:", err)
		os.Exit(1)
	}
}
