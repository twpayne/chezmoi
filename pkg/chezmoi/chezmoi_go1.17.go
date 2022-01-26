//go:build !go1.18
// +build !go1.18

package chezmoi

import (
	"bytes"
	"strings"
)

// CutBytes slices s around the first instance of sep, returning the text before
// and after sep. The found result reports whether sep appears in s. If sep does
// not appear in s, cut returns s, nil, false.
//
// CutBytes returns slices of the original slice s, not copies.
func CutBytes(s, sep []byte) (before, after []byte, found bool) {
	if i := bytes.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, nil, false
}

// CutString slices s around the first instance of sep, returning the text
// before and after sep. The found result reports whether sep appears in s. If
// sep does not appear in s, cut returns s, "", false.
func CutString(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}
