//go:build test

package chezmoitest

import (
	"io/fs"
	"strconv"
)

// mustParseFileMode parses s as a fs.FileMode and panics on any error.
func mustParseFileMode(s string) fs.FileMode {
	u, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		panic(err)
	}
	return fs.FileMode(uint32(u))
}
