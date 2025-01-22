package prefixset

import (
	"iter"
	"strings"
)

// lines is copied from the Go 1.24 source tree.
// Remove this once we upgrade to Go 1.24.
func lines(s string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for len(s) > 0 {
			var line string
			if i := strings.IndexByte(s, '\n'); i >= 0 {
				line, s = s[:i+1], s[i+1:]
			} else {
				line, s = s, ""
			}
			if !yield(line) {
				return
			}
		}
	}
}
