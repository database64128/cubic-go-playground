//go:build dragonfly || freebsd || linux || netbsd || openbsd || solaris

package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func socket(domain int, typ int, proto int) (fd int, err error) {
	fd, err = unix.Socket(domain, typ|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, proto)
	if err != nil {
		return 0, os.NewSyscallError("socket", err)
	}
	return fd, nil
}
