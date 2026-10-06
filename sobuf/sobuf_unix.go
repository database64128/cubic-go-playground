//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func run() int {
	fd, err := socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open UDP socket:", err)
		return 1
	}
	defer unix.Close(fd)

	if !printSocketBufferSizes(fd) {
		return 1
	}

	if size > 0 {
		fmt.Printf("\nSetting socket buffer sizes to %d\n\n", size)

		if !setSocketBufferSizes(fd, size, size) {
			return 1
		}

		if !printSocketBufferSizes(fd) {
			return 1
		}
	}

	return 0
}

func printSocketBufferSizes(fd int) bool {
	sndbuf, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get UDP socket send buffer size:", err)
		return false
	}

	rcvbuf, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get UDP socket receive buffer size:", err)
		return false
	}

	fmt.Printf("UDP socket %d:\n  Send buffer size:    %d\n  Receive buffer size: %d\n", fd, sndbuf, rcvbuf)
	return true
}

func setSocketBufferSizes(fd int, sndbuf, rcvbuf int) bool {
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF, sndbuf); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set UDP socket send buffer size:", err)
		return false
	}

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, rcvbuf); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set UDP socket receive buffer size:", err)
		return false
	}

	return true
}
