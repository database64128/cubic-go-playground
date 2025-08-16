//go:build darwin || dragonfly || freebsd || netbsd

package main

const sizeofMsghdr = 4 // int(unsafe.Sizeof(msghdr{}))

type msghdr struct {
	Msglen  uint16
	Version uint8
	Type    uint8
}

func (m *msghdr) isHdrlenOK() bool {
	return true
}

func (m *msghdr) addrsBuf(msgBuf []byte, hdrlen int) []byte {
	return msgBuf[hdrlen:]
}
