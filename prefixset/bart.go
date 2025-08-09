package prefixset

import (
	"net/netip"
	"strings"

	"github.com/gaissmai/bart"
)

// BARTFromText parses prefixes from text and builds a table.
func BARTFromText(text string) (*bart.Lite, error) {
	var s bart.Lite
	if err := bartInsertFromText(&s, text); err != nil {
		return nil, err
	}
	return &s, nil
}

func bartInsertFromText(s *bart.Lite, text string) error {
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
			return err
		}

		s.Insert(prefix)
	}

	return nil
}
