// Package archivetest provides useful functions for testing archives.
package archivetest

import (
	"io/fs"
)

// A Dir represents a directory.
type Dir struct {
	Entries map[string]any
	Perm    fs.FileMode
}

// A File represents a file.
type File struct {
	Contents []byte
	Perm     fs.FileMode
}

// A Symlink represents a symlink.
type Symlink struct {
	Target string
}
