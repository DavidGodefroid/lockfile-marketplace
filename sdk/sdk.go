// Package sdk is what a lockfile plugin's main calls. The plugin answers
// one yes/no question about the file it gets on fd 3; the sdk turns the
// answer into the protocol (stdout "1" or "0", exit 0) and turns any error
// or panic into a silent failure (nothing on stdout or stderr, exit 1),
// so no detail about the file can leave through an error message.
package sdk

import "os"

// InputFD is the descriptor the sandbox passes the file on.
const InputFD = 3

// Main runs answer on the input file and exits. It never returns.
func Main(answer func(f *os.File) (bool, error)) {
	defer func() {
		if recover() != nil {
			os.Exit(1)
		}
	}()
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
