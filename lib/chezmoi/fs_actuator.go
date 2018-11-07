package chezmoi

import (
	"os"

	"github.com/absfs/afero"
	"github.com/google/renameio"
	"github.com/pkg/errors"
)

// An FsActuator makes changes to an afero.Fs.
type FsActuator struct {
	afero.Fs
	dir string
}

// NewFsActuator returns an actuator that acts on fs.
func NewFsActuator(fs afero.Fs, targetDir string) *FsActuator {
	var dir string
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if _, ok := fs.(*afero.OsFs); ok {
		dir = renameio.TempDir(targetDir)
	}
	return &FsActuator{
		Fs:  fs,
		dir: dir,
	}
}

// WriteFile implements Actuator.WriteFile.
func (a *FsActuator) WriteFile(name string, contents []byte, mode os.FileMode, currentContents []byte) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if _, ok := a.Fs.(*afero.OsFs); ok {
		t, err := renameio.TempFile(a.dir, name)
		if err != nil {
			return err
		}
		defer t.Cleanup()
		if err := t.Chmod(mode); err != nil {
			return err
		}
		n, err := t.Write(contents)
		if err != nil {
			return err
		}
		if n != len(contents) {
			return errors.Errorf("%s: wrote %d bytes, want %d", name, n, len(contents))
		}
		return t.CloseAtomicallyReplace()
	}
	return afero.WriteFile(a.Fs, name, contents, mode)
}
