// Package harness is the conformance test every marketplace plugin runs:
// it builds the plugin, validates its manifest, runs it in the lockfile
// sandbox on each fixture listed in testdata/expected.toml (at least one
// per answer), then on random and truncated inputs, which must each give
// an answer or a clean failure, never a hang or a crash.
//
// A plugin's test package must call sandbox.Main() first in TestMain: the
// sandbox re-executes the test binary as its trampoline.
package harness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/DavidGodefroid/lockfile/pluginspec"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

// Fuzz sizes.
const (
	RandomInputs    = 50
	TruncationsEach = 50
	maxRandomSize   = 8 << 10
)

type expected struct {
	Case []Case `toml:"case"`
}

// Case is one fixture and the answer the plugin must give on it.
type Case struct {
	File   string `toml:"file"`
	Answer int    `toml:"answer"`
}

// Check runs Verify and fails t with its error.
func Check(t *testing.T, pluginDir string) {
	t.Helper()
	if err := Verify(pluginDir); err != nil {
		t.Fatal(err)
	}
}

// Verify runs the conformance checks on the plugin in pluginDir.
func Verify(pluginDir string) error {
	dir, err := filepath.Abs(pluginDir)
	if err != nil {
		return err
	}
	mb, err := os.ReadFile(filepath.Join(dir, "plugin.toml"))
	if err != nil {
		return err
	}
	m, err := pluginspec.ParseManifest(mb)
	if err != nil {
		return err
	}
	cases, err := readExpected(filepath.Join(dir, "testdata", "expected.toml"))
	if err != nil {
		return err
	}
	bin, cleanup, err := build(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx := context.Background()
	var errs []error
	var fixtures [][]byte
	for _, c := range cases {
		b, err := os.ReadFile(filepath.Join(dir, "testdata", c.File))
		if err != nil {
			return fmt.Errorf("fixture %s: %w", c.File, err)
		}
		fixtures = append(fixtures, b)
		o, rep, err := run(ctx, bin, b, m.Limits)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("%s: sandbox: %w", c.File, err))
		case o.Answer == nil:
			errs = append(errs, fmt.Errorf("%s: want %d, got no answer (%s, exit %d, signal %q)", c.File, c.Answer, o.Cause, rep.ExitCode, rep.Signal))
		case *o.Answer != (c.Answer == 1):
			errs = append(errs, fmt.Errorf("%s: want %d, got %v", c.File, c.Answer, *o.Answer))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return fuzz(ctx, bin, m, fixtures)
}

func readExpected(path string) ([]Case, error) {
	var e expected
	md, err := toml.DecodeFile(path, &e)
	if err != nil {
		return nil, err
	}
	if u := md.Undecoded(); len(u) > 0 {
		return nil, fmt.Errorf("%s: unknown keys %v", path, u)
	}
	seen := map[int]bool{}
	for _, c := range e.Case {
		if c.Answer != 0 && c.Answer != 1 {
			return nil, fmt.Errorf("%s: %s: answer must be 0 or 1", path, c.File)
		}
		if c.File == "" || filepath.Base(c.File) != c.File {
			return nil, fmt.Errorf("%s: fixture %q must be a file name in testdata", path, c.File)
		}
		seen[c.Answer] = true
	}
	var errs []error
	for _, a := range []int{0, 1} {
		if !seen[a] {
			errs = append(errs, fmt.Errorf("%s: needs at least one fixture with answer %d", path, a))
		}
	}
	return e.Case, errors.Join(errs...)
}

func build(dir string) (string, func(), error) {
	tmp, err := os.MkdirTemp("", "lockfile-harness-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmp) }
	out := filepath.Join(tmp, "plugin")
	cmd := exec.Command("go", "build", "-trimpath", "-o", out, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if b, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("build %s: %v\n%s", dir, err, b)
	}
	return out, cleanup, nil
}

func run(ctx context.Context, bin string, input []byte, l pluginspec.Limits) (pluginspec.Outcome, sandbox.Report, error) {
	in, err := os.CreateTemp("", "lockfile-harness-input-")
	if err != nil {
		return pluginspec.Outcome{}, sandbox.Report{}, err
	}
	defer os.Remove(in.Name())
	defer in.Close()
	if _, err := in.Write(input); err != nil {
		return pluginspec.Outcome{}, sandbox.Report{}, err
	}
	b, err := os.Open(bin)
	if err != nil {
		return pluginspec.Outcome{}, sandbox.Report{}, err
	}
	defer b.Close()
	return sandbox.Run(ctx, sandbox.Spec{Binary: b, Input: in, Limits: l})
}

// fuzz feeds random blobs and truncated fixtures; each must give an
// answer or exit with a status (a clean failure). A timeout, a signal or
// malformed output fails. It stops at the first failure.
func fuzz(ctx context.Context, bin string, m pluginspec.Manifest, fixtures [][]byte) error {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	rng := rand.New(rand.NewPCG(h.Sum64(), 1))
	type input struct {
		name string
		b    []byte
	}
	var inputs []input
	for i := 0; i < RandomInputs; i++ {
		b := make([]byte, rng.IntN(maxRandomSize))
		for j := range b {
			b[j] = byte(rng.Uint32())
		}
		inputs = append(inputs, input{fmt.Sprintf("random #%d (%d bytes)", i, len(b)), b})
	}
	for fi, f := range fixtures {
		for i := 0; i < TruncationsEach && len(f) > 0; i++ {
			n := rng.IntN(len(f))
			inputs = append(inputs, input{fmt.Sprintf("fixture #%d truncated to %d bytes", fi, n), bytes.Clone(f[:n])})
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	work := make(chan input)
	var mu sync.Mutex
	var first error
	fail := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if first == nil {
			first = err
			cancel()
		}
	}
	var wg sync.WaitGroup
	for w := 0; w < runtime.GOMAXPROCS(0); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for in := range work {
				o, rep, err := run(ctx, bin, in.b, m.Limits)
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					fail(fmt.Errorf("fuzz %s: sandbox: %w", in.name, err))
					continue
				}
				if o.Answer == nil && o.Cause != pluginspec.CauseExitCode {
					fail(fmt.Errorf("fuzz %s: %s (exit %d, signal %q)", in.name, o.Cause, rep.ExitCode, rep.Signal))
				}
			}
		}()
	}
	for _, in := range inputs {
		select {
		case work <- in:
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			break
		}
	}
	close(work)
	wg.Wait()
	return first
}
