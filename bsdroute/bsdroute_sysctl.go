//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"os"
	"slices"
	"unsafe"

	"golang.org/x/sys/unix"
)

//go:linkname sysctl syscall.sysctl
//go:noescape
func sysctl(mib []int32, old *byte, oldlen *uintptr, new *byte, newlen uintptr) (err error)

func sysctlGetBytes(mib []int32) (b []byte, err error) {
	for {
		var n uintptr
		if err := sysctl(mib, nil, &n, nil, 0); err != nil {
			return nil, os.NewSyscallError("sysctl", err)
		}
		b = slices.Grow(b, int(n))
		n = uintptr(cap(b))
		if err := sysctl(mib, unsafe.SliceData(b), &n, nil, 0); err != nil {
			if err == unix.ENOMEM {
				continue
			}
			return nil, os.NewSyscallError("sysctl", err)
		}
		return b[:n], nil
	}
}
