package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func run() int {
	fd, err := windows.WSASocket(windows.AF_INET, windows.SOCK_DGRAM, windows.IPPROTO_UDP, nil, 0, windows.WSA_FLAG_OVERLAPPED|windows.WSA_FLAG_NO_HANDLE_INHERIT)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open UDP socket:", err)
		return 1
	}
	defer windows.Closesocket(fd)

	if !printSocketBufferSizes(fd) {
		return 1
	}

	if size > 0 {
		fmt.Println()

		if !setSocketBufferSizes(fd, size, size) {
			return 1
		}

		fmt.Println()

		if !printSocketBufferSizes(fd) {
			return 1
		}
	}

	return 0
}

func printSocketBufferSizes(fd windows.Handle) bool {
	sndbuf, err := windows.GetsockoptInt(fd, windows.SOL_SOCKET, windows.SO_SNDBUF)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get UDP socket send buffer size:", err)
		return false
	}

	rcvbuf, err := windows.GetsockoptInt(fd, windows.SOL_SOCKET, windows.SO_RCVBUF)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get UDP socket receive buffer size:", err)
		return false
	}

	fmt.Printf("UDP socket %d:\n  Send buffer size:    %d\n  Receive buffer size: %d\n", fd, sndbuf, rcvbuf)
	return true
}

func setSocketBufferSizes(fd windows.Handle, sndbuf, rcvbuf int) bool {
	fmt.Printf("Setting SO_SNDBUF=%d, SO_RCVBUF=%d\n", sndbuf, rcvbuf)

	if err := windows.SetsockoptInt(fd, windows.SOL_SOCKET, windows.SO_SNDBUF, sndbuf); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set UDP socket send buffer size:", err)
		return false
	}

	if err := windows.SetsockoptInt(fd, windows.SOL_SOCKET, windows.SO_RCVBUF, rcvbuf); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set UDP socket receive buffer size:", err)
		return false
	}

	return true
}
