package chezmoi

import (
	"os"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
)

// An FSMutator makes changes to an vfs.FS.
type FSMutator struct {
	vfs.FS
	dir string
}

// NewFSMutator returns an mutator that acts on fs.
func NewFSMutator(fs vfs.FS, destDir string) *FSMutator {
	var dir string
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if fs == vfs.OSFS {
		dir = renameio.TempDir(destDir)
	}
	return &FSMutator{
		FS:  fs,
		dir: dir,
	}
}

// WriteFile implements Mutator.WriteFile.
func (a *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if a.FS == vfs.OSFS {
		t, err := renameio.TempFile(a.dir, name)
		if err != nil {
			return err
		}
		defer func() {
			_ = t.Cleanup()
		}()
		if err := t.Chmod(perm); err != nil {
			return err
		}
		if _, err := t.Write(data); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	return a.FS.WriteFile(name, data, perm)
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
