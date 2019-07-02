package chezmoi

import "os"

// A Mutator makes changes.
type Mutator interface {
    IsPrivate(file string, umask os.FileMode) bool
	Chmod(name string, mode os.FileMode) error
	Mkdir(name string, perm os.FileMode) error
	RemoveAll(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
	WriteFile(filename string, data []byte, perm os.FileMode, currData []byte) error
	WriteSymlink(oldname, newname string) error
}
