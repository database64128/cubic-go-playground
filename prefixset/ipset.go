package prefixset

import (
	"net/netip"

	"go4.org/netipx"
)

// IPSetFromText parses prefixes from the text and builds a prefix set.
func IPSetFromText(text string) (*netipx.IPSet, error) {
	var sb netipx.IPSetBuilder

	for line := range lines(text) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			if len(line) == 0 {
				continue
			}
			if line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
				if len(line) == 0 {
					continue
				}
			}
		}

		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, err
		}

		sb.AddPrefix(prefix)
	}

	return sb.IPSet()
}

// IPSetToText returns the text representation of the prefix set.
func IPSetToText(s *netipx.IPSet) []byte {
	prefixes := s.Prefixes()
	const typicalLineLength = len("ffff:ffff:ffff::/48\n")
	b := make([]byte, 0, len(prefixes)*typicalLineLength)
	for _, prefix := range prefixes {
		b = prefix.AppendTo(b)
		b = append(b, '\n')
	}
	return b
}
