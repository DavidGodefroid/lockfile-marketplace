// Package sdk is what a lockfile plugin's main calls. The plugin answers
// one yes/no question about the file it gets on fd 3; the sdk turns the
// answer into the protocol (stdout "1" or "0", exit 0) and turns any error
// or panic into a silent failure (nothing on stdout or stderr, exit 1),
// so no detail about the file can leave through an error message.
//
// A protocol-2 plugin (plugin.toml protocol = 2) may also declare inputs,
// values the human types for its question's ${name} placeholders, and a
// detail, a short text for the human only. It reads the values with
// Inputs and writes the detail to Detail; the agent never sees either.
package sdk

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"
)

// Descriptors the sandbox passes the plugin.
const (
	// InputFD is the file the plugin answers about, read-only.
	InputFD = 3
	// InputsFD holds the typed values (protocol 2, manifest inputs), a
	// sealed read-only memfd.
	InputsFD = 6
	// DetailFD is the write end of the detail pipe (protocol 2, manifest
	// detail = true).
	DetailFD = 7
)

// MaxDetail is how much of the detail the console keeps: the rest is
// drained unread. The Detail writer drops it already.
const MaxDetail = 4 << 10

// maxInputsJSON bounds what Inputs reads: 4 inputs of at most 1024 bytes,
// each escaped at worst to twice its size, with their names and quotes.
const maxInputsJSON = 16 << 10

const protocolEnv = "LOCKFILE_PLUGIN_PROTOCOL"

// errInputs is the one error of Inputs: it never quotes what fd 6 holds.
var errInputs = errors.New("sdk: no typed values")

// Option configures Main.
type Option func(*config)

type config struct {
	detail bool
}

// WithDetail tells Main that the plugin's manifest has detail = true, so
// the sandbox gave it a detail pipe on fd 7 and Detail writes there.
// Without it Detail discards everything and the sdk never touches fd 7:
// with no detail pipe the Go runtime may use that number itself.
func WithDetail() Option {
	return func(c *config) { c.detail = true }
}

var (
	mu         sync.Mutex
	detail     io.Writer = io.Discard
	detailFile *os.File  // kept so the GC never closes fd 7
)

// Main runs answer on the input file and exits. It never returns.
func Main(answer func(f *os.File) (bool, error), opts ...Option) {
	defer func() {
		if recover() != nil {
			os.Exit(1)
		}
	}()
	var c config
	for _, o := range opts {
		o(&c)
	}
	if c.detail && protocol() >= 2 {
		// Only a pipe on fd 7 is the sandbox's: should the manifest lack
		// detail, the number may be the runtime's own descriptor.
		detailFile = os.NewFile(DetailFD, "detail")
		if st, err := detailFile.Stat(); err == nil && st.Mode()&os.ModeNamedPipe != 0 {
			mu.Lock()
			detail = &cappedWriter{f: detailFile, left: MaxDetail}
			mu.Unlock()
		}
	}
	v, err := answer(os.NewFile(InputFD, "input"))
	if err != nil {
		os.Exit(1)
	}
	out := "0"
	if v {
		out = "1"
	}
	if _, err := os.Stdout.WriteString(out); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// protocol is the plugin protocol the sandbox announced, 1 when unset.
func protocol() int {
	p, err := strconv.Atoi(os.Getenv(protocolEnv))
	if err != nil || p < 1 {
		return 1
	}
	return p
}

// Inputs returns the values the human typed for the plugin's inputs, by
// input name. The sandbox passes them, under protocol 2 only, as the
// canonical JSON {"needle":"..."} (keys sorted) on fd 6; Inputs never
// reads argv, the environment or a path. Under protocol 1, or without
// a readable regular file on fd 6 holding a JSON object of strings, it
// returns an error that never quotes a value: return it from the answer
// function and the plugin fails silently.
//
// The console already checked each value against its input's kind and
// max_bytes before the run; anything else the question needs (a value
// the question cannot use) is the plugin's own check. A plugin uses the
// values only for its stated question and never writes them anywhere,
// not even to Detail.
func Inputs() (map[string]string, error) {
	return readInputs()
}

var (
	inputsFile *os.File // kept so the GC never closes fd 6
	readInputs = sync.OnceValues(func() (map[string]string, error) {
		if protocol() < 2 {
			return nil, errInputs
		}
		inputsFile = os.NewFile(InputsFD, "inputs")
		st, err := inputsFile.Stat()
		if err != nil || !st.Mode().IsRegular() {
			return nil, errInputs
		}
		b, err := io.ReadAll(io.NewSectionReader(inputsFile, 0, maxInputsJSON+1))
		if err != nil || len(b) > maxInputsJSON {
			return nil, errInputs
		}
		var m map[string]string
		if json.Unmarshal(b, &m) != nil || m == nil {
			clear(b)
			return nil, errInputs
		}
		clear(b)
		return m, nil
	})
)

// Detail returns where the plugin writes its detail: what the human needs
// to act on the answer ("found on pages 4512, 12307"), shown on the
// console with clear results on, never to the agent nor in the audit. It
// is the fd 7 pipe when Main got WithDetail under protocol 2, a writer
// that discards everything otherwise. Only the first MaxDetail bytes are
// kept. Writes never fail, so a detail can never change the answer. The
// detail must never echo a typed value.
func Detail() io.Writer {
	mu.Lock()
	defer mu.Unlock()
	return detail
}

// cappedWriter writes the first left bytes to f and drops the rest and
// every write error.
type cappedWriter struct {
	mu   sync.Mutex
	f    *os.File
	left int
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.left > 0 {
		b := p[:min(len(p), w.left)]
		w.left -= len(b)
		if _, err := w.f.Write(b); err != nil {
			w.left = 0
		}
	}
	return len(p), nil
}
