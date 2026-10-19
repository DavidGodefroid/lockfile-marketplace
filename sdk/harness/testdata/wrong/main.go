package main

import (
	"os"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

// Answers 1 to everything.
func main() {
	sdk.Main(func(*os.File) (bool, error) { return true, nil })
}
