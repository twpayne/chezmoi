package chezmoi

import "os"

// An Mutator makes changes.
type Mutator interface {
	Chmod(name string, mode os.FileMode) error
	Mkdir(name string, perm os.FileMode) error
	RemoveAll(name string) error
	Rename(oldpath, newpath string) error
	WriteFile(filename string, data []byte, perm os.FileMode, currData []byte) error
	WriteSymlink(oldname, newname string) error
}
