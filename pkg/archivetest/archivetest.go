package archivetest

import (
	"io/fs"
	"sort"
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

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
