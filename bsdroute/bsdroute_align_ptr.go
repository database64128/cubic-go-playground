//go:build dragonfly || freebsd || openbsd || solaris

package main

import "golang.org/x/sys/unix"

const rtaAlignTo = unix.SizeofPtr
