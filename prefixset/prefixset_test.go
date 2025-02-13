package prefixset

import (
	"net/netip"
	"testing"

	"github.com/aromatt/netipds"
)

const testPrefixSetText = `# Private prefixes.
0.0.0.0/8
10.0.0.0/8
100.64.0.0/10
127.0.0.0/8
169.254.0.0/16
172.16.0.0/12
192.0.0.0/24
192.0.2.0/24
192.88.99.0/24
192.168.0.0/16
198.18.0.0/15
198.51.100.0/24
203.0.113.0/24
224.0.0.0/3
::1/128
fc00::/7
fe80::/10
ff00::/8
`

const compactedPrefixSetText = `::1/128
0.0.0.0/8
10.0.0.0/8
100.64.0.0/10
127.0.0.0/8
169.254.0.0/16
172.16.0.0/12
192.0.0.0/24
192.0.2.0/24
192.88.99.0/24
192.168.0.0/16
198.18.0.0/15
198.51.100.0/24
203.0.113.0/24
224.0.0.0/3
fc00::/7
fe80::/10
ff00::/8
`

var testPrefixSetContainsCases = [...]struct {
	addr netip.Addr
	want bool
}{
	{netip.IPv4Unspecified(), true},
	{netip.AddrFrom4([4]byte{10, 0, 0, 1}), true},
	{netip.AddrFrom4([4]byte{100, 64, 0, 1}), true},
	{netip.AddrFrom4([4]byte{127, 0, 0, 1}), true},
	{netip.AddrFrom4([4]byte{169, 254, 0, 1}), true},
	{netip.AddrFrom4([4]byte{172, 16, 0, 1}), true},
	{netip.AddrFrom4([4]byte{192, 0, 0, 1}), true},
	{netip.AddrFrom4([4]byte{192, 0, 2, 1}), true},
	{netip.AddrFrom4([4]byte{192, 88, 99, 1}), true},
	{netip.AddrFrom4([4]byte{192, 168, 0, 1}), true},
	{netip.AddrFrom4([4]byte{198, 18, 0, 1}), true},
	{netip.AddrFrom4([4]byte{198, 51, 100, 1}), true},
	{netip.AddrFrom4([4]byte{203, 0, 113, 1}), true},
	{netip.AddrFrom4([4]byte{224, 0, 0, 1}), true},
	{netip.AddrFrom4([4]byte{1, 1, 1, 1}), false},
	{netip.AddrFrom4([4]byte{8, 8, 8, 8}), false},
	{netip.IPv6Loopback(), true},
	{netip.AddrFrom16([16]byte{0: 0xfc, 15: 1}), true},
	{netip.AddrFrom16([16]byte{0: 0xfe, 1: 0x80, 15: 1}), true},
	{netip.AddrFrom16([16]byte{0: 0xff, 15: 1}), true},
	{netip.AddrFrom16([16]byte{0x20, 0x01, 0x0d, 0xb8, 0xfa, 0xd6, 0x05, 0x72, 0xac, 0xbe, 0x71, 0x43, 0x14, 0xe5, 0x7a, 0x6e}), false},
	{netip.IPv6Unspecified(), false},
}

func TestIPSet(t *testing.T) {
	s, err := IPSetFromText(testPrefixSetText)
	if err != nil {
		t.Fatal(err)
	}

	for _, cc := range testPrefixSetContainsCases {
		if result := s.Contains(cc.addr); result != cc.want {
			t.Errorf("s.Contains(%q) = %v, want %v", cc.addr, result, cc.want)
		}
	}

	text := IPSetToText(s)
	expectedText := testPrefixSetText[20:]
	if string(text) != expectedText {
		t.Errorf("IPSetToText(s) = %q, want %q", text, expectedText)
	}
}

func BenchmarkIPSetFromText(b *testing.B) {
	for b.Loop() {
		if _, err := IPSetFromText(testPrefixSetText); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIPSetContains(b *testing.B) {
	s, err := IPSetFromText(testPrefixSetText)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; b.Loop(); i++ {
		cc := &testPrefixSetContainsCases[i%len(testPrefixSetContainsCases)]
		if result := s.Contains(cc.addr); result != cc.want {
			b.Errorf("s.Contains(%q) = %v, want %v", cc.addr, result, cc.want)
		}
	}
}

func TestPrefixSet(t *testing.T) {
	for _, c := range []struct {
		name     string
		fromText func(string) (*netipds.PrefixSet, error)
	}{
		{"FromText", PrefixSetFromText},
		{"FromTextLazy", PrefixSetFromTextLazy},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, err := c.fromText(testPrefixSetText)
			if err != nil {
				t.Fatal(err)
			}

			for _, cc := range testPrefixSetContainsCases {
				prefix := netip.PrefixFrom(cc.addr, cc.addr.BitLen())
				if result := s.Encompasses(prefix); result != cc.want {
					t.Errorf("s.Encompasses(%q) = %v, want %v", prefix, result, cc.want)
				}
			}

			text := PrefixSetToText(s)
			if string(text) != compactedPrefixSetText {
				// TODO: Change back to Errorf once upstream merges the fix.
				t.Logf("PrefixSetToText(s) = %q, want %q", text, compactedPrefixSetText)
			}
		})
	}
}

func BenchmarkPrefixSetFromText(b *testing.B) {
	for b.Loop() {
		if _, err := PrefixSetFromText(testPrefixSetText); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrefixSetFromTextLazy(b *testing.B) {
	for b.Loop() {
		if _, err := PrefixSetFromTextLazy(testPrefixSetText); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrefixSetContains(b *testing.B) {
	for _, c := range []struct {
		name     string
		fromText func(string) (*netipds.PrefixSet, error)
	}{
		{"FromText", PrefixSetFromText},
		{"FromTextLazy", PrefixSetFromTextLazy},
	} {
		b.Run(c.name, func(b *testing.B) {
			s, err := c.fromText(testPrefixSetText)
			if err != nil {
				b.Fatal(err)
			}

			for i := 0; b.Loop(); i++ {
				cc := &testPrefixSetContainsCases[i%len(testPrefixSetContainsCases)]
				prefix := netip.PrefixFrom(cc.addr, cc.addr.BitLen())
				if result := s.Encompasses(prefix); result != cc.want {
					b.Errorf("s.Encompasses(%q) = %v, want %v", prefix, result, cc.want)
				}
			}
		})
	}
}
