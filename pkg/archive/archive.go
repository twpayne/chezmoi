package archive

import "io/fs"

// A Dir represents a directory.
type Dir struct {
	Perm    fs.FileMode
	Entries map[string]interface{}
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
