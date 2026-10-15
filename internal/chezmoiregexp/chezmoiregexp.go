// Package chezmoiregexp wraps common functions from the standard regexp
// library.
package chezmoiregexp

import (
	"regexp"
	"sync"
)

// MustCompileRegexpOnce returns a regexp that is compiled when needed. The
// function will panic if the regexp is invalid.
func MustCompileRegexpOnce(s string) func() *regexp.Regexp {
	return sync.OnceValue(func() *regexp.Regexp {
		return regexp.MustCompile(s)
	})
}
