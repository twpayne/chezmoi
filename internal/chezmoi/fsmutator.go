package chezmoi

import (
	"os"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
)

// An FSMutator makes changes to an vfs.FS.
type FSMutator struct {
	vfs.FS
	devCache     map[string]uint // devCache maps directories to device numbers.
	tempDirCache map[uint]string // tempDir maps device numbers to renameio temporary directories.
}

// NewFSMutator returns an mutator that acts on fs.
func NewFSMutator(fs vfs.FS) *FSMutator {
	return &FSMutator{
		FS:           fs,
		devCache:     make(map[string]uint),
		tempDirCache: make(map[uint]string),
	}
}

// WriteSymlink implements Mutator.WriteSymlink.
func (a *FSMutator) WriteSymlink(oldname, newname string) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if a.FS == vfs.OSFS {
		return renameio.Symlink(oldname, newname)
	}
	if err := a.FS.RemoveAll(newname); err != nil && !os.IsNotExist(err) {
		return err
	}
	return a.FS.Symlink(oldname, newname)
}
