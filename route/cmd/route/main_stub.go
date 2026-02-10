//go:build !darwin && !dragonfly && !freebsd && !netbsd && !openbsd && !windows

package main

import (
	"os"
)

func main() {
	os.Stderr.WriteString("Unsupported platform\n")
	os.Exit(1)
}
