package main

import (
	"os"
	"time"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// Answers fixtures, but hangs on input it does not expect.
func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		b := make([]byte, 64)
		n, _ := f.Read(b)
		switch string(b[:n]) {
		case "Yes, synthetic.\n":
			return true, nil
		case "No, synthetic.\n":
			return false, nil
		}
		time.Sleep(time.Hour)
		return false, nil
	})
}
