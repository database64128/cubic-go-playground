//go:build openbsd

package main

const sizeofMsghdr = 6 // int(unsafe.Sizeof(msghdr{}))

type msghdr struct {
	Msglen  uint16
	Version uint8
	Type    uint8
	Hdrlen  uint16
}

func (m *msghdr) isHdrlenOK() bool {
	return m.Hdrlen <= m.Msglen
}

func (m *msghdr) addrsBuf(msgBuf []byte, _ int) []byte {
	return msgBuf[m.Hdrlen:]
}
