//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

func rtaAlign(n int) int {
	return (n + rtaAlignTo - 1) & ^(rtaAlignTo - 1)
}
