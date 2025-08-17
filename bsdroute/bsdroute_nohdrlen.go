//go:build darwin || dragonfly || freebsd || netbsd

package main

const sizeofMsghdr = 4 // int(unsafe.Sizeof(msghdr{}))

type msghdr struct {
	Msglen  uint16
	Version uint8
	Type    uint8
}

func (*msghdr) hdrlen() uint16 {
	panic("unreachable")
}

func (*msghdr) addrsBuf(msgBuf []byte, hdrlen int) ([]byte, bool) {
	return msgBuf[hdrlen:], true
}
