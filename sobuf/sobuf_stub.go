//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !windows

package main

import (
	"fmt"
	"os"
)

func run() int {
	fmt.Fprintln(os.Stderr, "Unsupported platform")
	return 1
}
