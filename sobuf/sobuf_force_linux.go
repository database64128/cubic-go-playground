package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func setSocketBufferSizesForce(fd int, sndbuf, rcvbuf int) bool {
	fmt.Printf("Setting SO_SNDBUFFORCE=%d, SO_RCVBUFFORCE=%d\n", sndbuf, rcvbuf)

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUFFORCE, sndbuf); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set UDP socket send buffer size:", err)
		return false
	}

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUFFORCE, rcvbuf); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set UDP socket receive buffer size:", err)
		return false
	}

	return true
}
