package prefixset

import (
	"net/netip"
	"strings"

	"github.com/aromatt/netipds"
)

// PrefixSetFromText parses prefixes from the text and builds a prefix set.
func PrefixSetFromText(text string) (*netipds.PrefixSet, error) {
	return prefixSetFromText(text, netipds.PrefixSetBuilder{})
}

// PrefixSetFromTextLazy is like [PrefixSetFromText] but uses a lazy builder.
func PrefixSetFromTextLazy(text string) (*netipds.PrefixSet, error) {
	return prefixSetFromText(text, netipds.PrefixSetBuilder{Lazy: true})
}

func prefixSetFromText(text string, sb netipds.PrefixSetBuilder) (*netipds.PrefixSet, error) {
	for line := range strings.Lines(text) {
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

		if err = sb.Add(prefix); err != nil {
			return nil, err
		}
	}

	return sb.PrefixSet(), nil
}

// PrefixSetToText returns the text representation of the prefix set.
func PrefixSetToText(s *netipds.PrefixSet) []byte {
	const typicalLineLength = len("ffff:ffff:ffff::/48\n")
	b := make([]byte, 0, s.Size()*typicalLineLength)
	for prefix := range s.AllCompact() {
		b = prefix.AppendTo(b)
		b = append(b, '\n')
	}
	return b
}
