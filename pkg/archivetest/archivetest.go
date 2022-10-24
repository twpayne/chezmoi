// Package archivetest provides useful functions for testing archives.
package archivetest

import (
	"io/fs"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// A Dir represents a directory.
type Dir struct {
	Perm    fs.FileMode
	Entries map[string]any
}

// A File represents a file.
type File struct {
	Perm     fs.FileMode
	Contents []byte
}

// A Symlink represents a symlink.
type Symlink struct {
	Target string
}

func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}
