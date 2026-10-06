//go:build aix || darwin || dragonfly || freebsd || netbsd || openbsd || solaris

package main

import (
	"fmt"
	"os"
)

func setSocketBufferSizesForce(int, int, int) bool {
	fmt.Fprintln(os.Stderr, "SO_{SND,RCV}BUFFORCE not available on this platform")
	return false
}
