package ip

import (
	"net/netip"
	"unsafe"
)

// AddrPortMappedEqual returns whether the two addresses point to the same endpoint.
// An IPv4 address and an IPv4-mapped IPv6 address pointing to the same endpoint are considered equal.
// For example, 1.1.1.1:53 and [::ffff:1.1.1.1]:53 are considered equal.
func AddrPortMappedEqual(l, r netip.AddrPort) bool {
	if l == r {
		return true
	}
	return l.Port() == r.Port() && l.Addr().Unmap() == r.Addr().Unmap()
}

type addrPortHeader struct {
	ip   [16]byte
	z    unsafe.Pointer
	port uint16
}

func AddrPortMappedEqualUnsafe(l, r netip.AddrPort) bool {
	lp := (*addrPortHeader)(unsafe.Pointer(&l))
	rp := (*addrPortHeader)(unsafe.Pointer(&r))
	return lp.ip == rp.ip && lp.port == rp.port
}

// AddrPortv4Mappedv6 converts an IPv4 address to an IPv4-mapped IPv6 address.
// This function does nothing if addrPort is an IPv6 address.
func AddrPortv4Mappedv6(addrPort netip.AddrPort) netip.AddrPort {
	if addrPort.Addr().Is4() {
		addr6 := addrPort.Addr().As16()
		ip := netip.AddrFrom16(addr6)
		port := addrPort.Port()
		return netip.AddrPortFrom(ip, port)
	}
	return addrPort
}

func AddrPortv4Mappedv6Unsafe(addrPort netip.AddrPort) netip.AddrPort {
	if addrPort.Addr().Is4() {
		app := (*addrPortHeader)(unsafe.Pointer(&addrPort))
		app.z = unsafe.Add(app.z, 3*unsafe.Sizeof(uintptr(0)))
	}
	return addrPort
}
