//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

func rtaAlign(n uint8) uint8 {
	if n == 0 {
		return rtaAlignTo
	}
	return (n + rtaAlignTo - 1) & ^uint8(rtaAlignTo-1)
}
