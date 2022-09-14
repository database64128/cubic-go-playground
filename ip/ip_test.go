package ip

import (
	"net/netip"
	"testing"
	"unsafe"
)

var (
	addrPort4    = netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), 1080)
	addrPort4in6 = netip.AddrPortFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 127, 0, 0, 1}), 1080)
)

func BenchmarkAddrPortMappedEqual(b *testing.B) {
	b.Run("Equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortMappedEqual(addrPort4, addrPort4)
		}
	})

	b.Run("NotEqual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortMappedEqual(addrPort4, addrPort4in6)
		}
	})
}

func BenchmarkAddrPortMappedEqualUnsafe(b *testing.B) {
	b.Run("Equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortMappedEqualUnsafe(addrPort4, addrPort4)
		}
	})

	b.Run("NotEqual", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortMappedEqualUnsafe(addrPort4, addrPort4in6)
		}
	})
}

func BenchmarkAddrPortv4Mappedv6(b *testing.B) {
	b.Run("Is4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortv4Mappedv6(addrPort4)
		}
	})

	b.Run("Is4In6", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortv4Mappedv6(addrPort4in6)
		}
	})
}

func BenchmarkAddrPortv4Mappedv6Unsafe(b *testing.B) {
	b.Run("Is4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortv4Mappedv6Unsafe(addrPort4)
		}
	})

	b.Run("Is4In6", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			AddrPortv4Mappedv6Unsafe(addrPort4in6)
		}
	})
}

func TestAddrPortv4Mappedv6Unsafe(t *testing.T) {
	app := (*addrPortHeader)(unsafe.Pointer(&addrPort4))
	t.Logf("addrPort4.z: %p", app.z)
	app = (*addrPortHeader)(unsafe.Pointer(&addrPort4in6))
	t.Logf("addrPort4in6.z: %p", app.z)

	if ap := AddrPortv4Mappedv6Unsafe(addrPort4); ap != addrPort4in6 {
		t.Errorf("AddrPortv4Mappedv6Unsafe(%s) returned %s, expected %s.", addrPort4, ap, addrPort4in6)
	}

	if ap := AddrPortv4Mappedv6Unsafe(addrPort4in6); ap != addrPort4in6 {
		t.Errorf("AddrPortv4Mappedv6Unsafe(%s) returned %s, expected %s.", addrPort4in6, ap, addrPort4in6)
	}
}
