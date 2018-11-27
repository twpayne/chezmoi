package chezmoi

import (
	"os"

	"github.com/google/renameio"
	"github.com/twpayne/go-vfs"
)

// An FSActuator makes changes to an vfs.FS.
type FSActuator struct {
	vfs.FS
	dir string
}

// NewFSActuator returns an actuator that acts on fs.
func NewFSActuator(fs vfs.FS, targetDir string) *FSActuator {
	var dir string
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if fs == vfs.OSFS {
		dir = renameio.TempDir(targetDir)
	}
	return &FSActuator{
		FS:  fs,
		dir: dir,
	}
}

// WriteFile implements Actuator.WriteFile.
func (a *FSActuator) WriteFile(name string, contents []byte, perm os.FileMode, currentContents []byte) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if a.FS == vfs.OSFS {
		t, err := renameio.TempFile(a.dir, name)
		if err != nil {
			return err
		}
		defer t.Cleanup()
		if err := t.Chmod(perm); err != nil {
			return err
		}
		if _, err := t.Write(contents); err != nil {
			return err
		}
		return t.CloseAtomicallyReplace()
	}
	return a.FS.WriteFile(name, contents, perm)
}
