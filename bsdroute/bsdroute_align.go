//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

func rtaAlign(n int) int {
	if n == 0 {
		return rtaAlignTo
	}
	return (n + rtaAlignTo - 1) & ^(rtaAlignTo - 1)
}
