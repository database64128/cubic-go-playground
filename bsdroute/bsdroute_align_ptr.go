//go:build dragonfly || freebsd || openbsd

package main

import "golang.org/x/sys/unix"

const rtaAlignTo = unix.SizeofPtr
