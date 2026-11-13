// Package harness is the conformance test every marketplace plugin runs:
// it builds the plugin, validates its manifest, runs it in the lockfile
// sandbox on each fixture listed in testdata/expected.toml (at least one
// per answer), then on random and truncated inputs, which must each give
// an answer or a clean failure, never a hang or a crash.
//
// A protocol-2 plugin with inputs gets, on every run, the values of its
// case (expected.toml: inputs = { needle = "..." }) the way the console
// passes them: canonical JSON on fd 6, never argv or env. Every case must
// give each input a value that fits it, the fuzzing also draws random
// values within each input's kind and max_bytes, and a detail (fd 7) that
// echoes a value fails the check.
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
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/BurntSushi/toml"

	"github.com/DavidGodefroid/lockfile/pluginspec"
	"github.com/DavidGodefroid/lockfile/sandbox"
)

// Fuzz sizes.
const (
	RandomInputs    = 50
	TruncationsEach = 50
	// ValuesEach is, for a plugin with inputs, how many runs each intact
	// fixture gets with random values.
	ValuesEach    = 20
	maxRandomSize = 8 << 10
)

type expected struct {
	Case []Case `toml:"case"`
}

// Case is one fixture, the values typed for the plugin's inputs (protocol
// 2), and the answer the plugin must give on it.
type Case struct {
	File   string            `toml:"file"`
	Inputs map[string]string `toml:"inputs"`
	Answer int               `toml:"answer"`
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
	m, err := readManifest(dir)
	if err != nil {
		return err
	}
	cases, err := readExpected(filepath.Join(dir, "testdata", "expected.toml"), m)
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
	var fixtures []fixture
	for _, c := range cases {
		b, err := os.ReadFile(filepath.Join(dir, "testdata", c.File))
		if err != nil {
			return fmt.Errorf("fixture %s: %w", c.File, err)
		}
		fixtures = append(fixtures, fixture{b, c.Inputs})
		var detail bytes.Buffer
		o, rep, err := run(ctx, bin, b, m, c.Inputs, &detail)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("%s: sandbox: %w", c.File, err))
		case o.Answer == nil:
			errs = append(errs, fmt.Errorf("%s: want %d, got no answer (%s, exit %d, signal %q)", c.File, c.Answer, o.Cause, rep.ExitCode, rep.Signal))
		case *o.Answer != (c.Answer == 1):
			errs = append(errs, fmt.Errorf("%s: want %d, got %v", c.File, c.Answer, *o.Answer))
		}
		if n := echoed(detail.Bytes(), c.Inputs); n != "" {
			errs = append(errs, fmt.Errorf("%s: the detail echoes the value of input %s", c.File, n))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return fuzz(ctx, bin, m, fixtures)
}

// CheckNoAnswer runs NoAnswer and fails t with its error.
func CheckNoAnswer(t *testing.T, pluginDir string, inputs map[string][]byte) {
	t.Helper()
	if err := NoAnswer(pluginDir, inputs); err != nil {
		t.Fatal(err)
	}
}

// NoAnswer checks that the plugin gives no answer, and fails cleanly (exit
// status, no hang, no signal), on each input: the cases its question does
// not cover, such as a file that is not a PDF. A plugin with inputs gets
// the values of the first case of its expected.toml.
func NoAnswer(pluginDir string, inputs map[string][]byte) error {
	dir, err := filepath.Abs(pluginDir)
	if err != nil {
		return err
	}
	m, err := readManifest(dir)
	if err != nil {
		return err
	}
	var vals map[string]string
	if len(m.Inputs) > 0 {
		cases, err := readExpected(filepath.Join(dir, "testdata", "expected.toml"), m)
		if err != nil {
			return err
		}
		vals = cases[0].Inputs
	}
	bin, cleanup, err := build(dir)
	if err != nil {
		return err
	}
	defer cleanup()
	var errs []error
	for name, in := range inputs {
		o, rep, err := run(context.Background(), bin, in, m, vals, nil)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("%s: sandbox: %w", name, err))
		case o.Answer != nil:
			errs = append(errs, fmt.Errorf("%s: want no answer, answered %v", name, *o.Answer))
		case o.Cause != pluginspec.CauseExitCode:
			errs = append(errs, fmt.Errorf("%s: want a clean failure, got %s (exit %d, signal %q)", name, o.Cause, rep.ExitCode, rep.Signal))
		}
	}
	return errors.Join(errs...)
}

func readManifest(dir string) (pluginspec.Manifest, error) {
	mb, err := os.ReadFile(filepath.Join(dir, "plugin.toml"))
	if err != nil {
		return pluginspec.Manifest{}, err
	}
	return pluginspec.ParseManifest(mb)
}

func readExpected(path string, m pluginspec.Manifest) ([]Case, error) {
	var e expected
	md, err := toml.DecodeFile(path, &e)
	if err != nil {
		return nil, err
	}
	if u := md.Undecoded(); len(u) > 0 {
		return nil, fmt.Errorf("%s: unknown keys %v", path, u)
	}
	seen := map[int]bool{}
	var errs []error
	for _, c := range e.Case {
		if c.Answer != 0 && c.Answer != 1 {
			return nil, fmt.Errorf("%s: %s: answer must be 0 or 1", path, c.File)
		}
		if c.File == "" || filepath.Base(c.File) != c.File {
			return nil, fmt.Errorf("%s: fixture %q must be a file name in testdata", path, c.File)
		}
		seen[c.Answer] = true
		errs = append(errs, checkValues(path, c, m)...)
	}
	for _, a := range []int{0, 1} {
		if !seen[a] {
			errs = append(errs, fmt.Errorf("%s: needs at least one fixture with answer %d", path, a))
		}
	}
	return e.Case, errors.Join(errs...)
}

// checkValues checks that a case gives each of the manifest's inputs a
// value that fits it, and nothing else. Its errors never quote a value.
func checkValues(path string, c Case, m pluginspec.Manifest) []error {
	var errs []error
	for _, in := range m.Inputs {
		v, ok := c.Inputs[in.Name]
		if !ok {
			errs = append(errs, fmt.Errorf("%s: %s: no value for input %s", path, c.File, in.Name))
		} else if in.CheckValue([]byte(v)) != nil {
			errs = append(errs, fmt.Errorf("%s: %s: the value of input %s does not fit its kind %s and max_bytes %d", path, c.File, in.Name, in.Kind, in.MaxBytes))
		}
	}
	for _, n := range sortedKeys(c.Inputs) {
		if _, ok := m.Input(n); !ok {
			errs = append(errs, fmt.Errorf("%s: %s: %s is not an input of the manifest", path, c.File, n))
		}
	}
	return errs
}

// echoed returns the name of an input whose value appears in detail.
func echoed(detail []byte, vals map[string]string) string {
	for _, n := range sortedKeys(vals) {
		if v := vals[n]; v != "" && bytes.Contains(detail, []byte(v)) {
			return n
		}
	}
	return ""
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// canonical is the JSON the console puts on fd 6: {"needle":"..."}, keys
// sorted, only '"' and '\' escaped (pluginspec.Input.CheckValue leaves no
// control character to escape).
func canonical(vals map[string]string) []byte {
	b := []byte{'{'}
	for i, k := range sortedKeys(vals) {
		if i > 0 {
			b = append(b, ',')
		}
		b = appendJSONString(b, k)
		b = append(b, ':')
		b = appendJSONString(b, vals[k])
	}
	return append(b, '}')
}

func appendJSONString(b []byte, s string) []byte {
	b = append(b, '"')
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\\' {
			b = append(b, '\\')
		}
		b = append(b, s[i])
	}
	return append(b, '"')
}

func build(dir string) (string, func(), error) {
	tmp, err := os.MkdirTemp("", "lockfile-harness-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmp) }
	out := filepath.Join(tmp, "plugin")
	cmd := exec.Command("go", "build", "-trimpath", "-buildvcs=false", "-o", out, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if b, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("build %s: %v\n%s", dir, err, b)
	}
	return out, cleanup, nil
}

// run runs the plugin as the console does: its manifest protocol, vals on
// fd 6 when the manifest has inputs, and fd 7 when it has detail, whose
// content (at most pluginspec.MaxDetail bytes) goes to detail (nil:
// dropped).
func run(ctx context.Context, bin string, input []byte, m pluginspec.Manifest, vals map[string]string, detail io.Writer) (pluginspec.Outcome, sandbox.Report, error) {
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
	spec := sandbox.Spec{Binary: b, Input: in, Limits: m.Limits, Protocol: m.Protocol}
	if len(m.Inputs) > 0 {
		spec.Inputs = canonical(vals)
	}
	if m.Detail {
		if detail == nil {
			detail = io.Discard
		}
		spec.Detail = detail
	}
	return sandbox.Run(ctx, spec)
}

type fixture struct {
	b    []byte
	vals map[string]string
}

// fuzz feeds random blobs and truncated fixtures, and for a plugin with
// inputs the intact fixtures with random values; each must give an answer
// or exit with a status (a clean failure). A timeout, a signal or
// malformed output fails. It stops at the first failure.
func fuzz(ctx context.Context, bin string, m pluginspec.Manifest, fixtures []fixture) error {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	rng := rand.New(rand.NewPCG(h.Sum64(), 1))
	type input struct {
		name string
		b    []byte
		vals map[string]string
	}
	// values draws random values for the inputs, falling back on the
	// fixture's own when none can be drawn (an email input too short).
	values := func(own map[string]string) map[string]string {
		if len(m.Inputs) == 0 {
			return nil
		}
		out := map[string]string{}
		for _, in := range m.Inputs {
			v, ok := randomValue(rng, in)
			if !ok {
				v = own[in.Name]
			}
			out[in.Name] = v
		}
		return out
	}
	var own map[string]string
	if len(fixtures) > 0 {
		own = fixtures[0].vals
	}
	var inputs []input
	for i := 0; i < RandomInputs; i++ {
		b := make([]byte, rng.IntN(maxRandomSize))
		for j := range b {
			b[j] = byte(rng.Uint32())
		}
		inputs = append(inputs, input{fmt.Sprintf("random #%d (%d bytes)", i, len(b)), b, values(own)})
	}
	for fi, f := range fixtures {
		for i := 0; i < TruncationsEach && len(f.b) > 0; i++ {
			n := rng.IntN(len(f.b))
			inputs = append(inputs, input{fmt.Sprintf("fixture #%d truncated to %d bytes", fi, n), bytes.Clone(f.b[:n]), f.vals})
		}
		for i := 0; i < ValuesEach && len(m.Inputs) > 0; i++ {
			inputs = append(inputs, input{fmt.Sprintf("fixture #%d with random values #%d", fi, i), f.b, values(f.vals)})
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
				o, rep, err := run(ctx, bin, in.b, m, in.vals, nil)
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

// Runes a random text value is drawn from, beyond printable ASCII: the
// characters JSON escapes, and multi-byte UTF-8.
var textRunes = []rune{'"', '\\', '/', 'é', 'ß', '€', '漢', '😀', ' ', ' '}

const (
	emailLocal  = "abcdefghijklmnopqrstuvwxyz0123456789._+-"
	emailDomain = "abcdefghijklmnopqrstuvwxyz0123456789-"
)

// randomValue draws a value that fits in (pluginspec.Input.CheckValue);
// false when it finds none.
func randomValue(rng *rand.Rand, in pluginspec.Input) (string, bool) {
	for try := 0; try < 16; try++ {
		var b []byte
		switch in.Kind {
		case pluginspec.KindDigits:
			n := 1 + rng.IntN(in.MaxBytes)
			for range n {
				b = append(b, byte('0'+rng.IntN(10)))
			}
		case pluginspec.KindEmail:
			// local@domain.tld, each part at least one byte.
			if in.MaxBytes < 5 {
				return "", false
			}
			n := 5 + rng.IntN(in.MaxBytes-4)
			l := 1 + rng.IntN(n-4)
			d := 1 + rng.IntN(n-3-l)
			t := n - 2 - l - d
			b = append(b, pick(rng, emailLocal, l)...)
			b = append(b, '@')
			b = append(b, pick(rng, emailDomain, d)...)
			b = append(b, '.')
			b = append(b, pick(rng, "abcdefghijklmnopqrstuvwxyz", t)...)
		default:
			n := 1 + rng.IntN(in.MaxBytes)
			for len(b) < n {
				r := rune(0x20 + rng.IntN(0x7f-0x20))
				if rng.IntN(4) == 0 {
					r = textRunes[rng.IntN(len(textRunes))]
				}
				b = append(b, string(r)...)
			}
			for len(b) > n {
				// Drop whole runes until it fits.
				_, size := utf8.DecodeLastRune(b)
				b = b[:len(b)-size]
			}
		}
		if len(b) > 0 && in.CheckValue(b) == nil {
			return string(b), true
		}
	}
	return "", false
}

func pick(rng *rand.Rand, alphabet string, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[rng.IntN(len(alphabet))]
	}
	return b
}
