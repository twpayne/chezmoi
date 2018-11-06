package chezmoi

import (
	"os"

	"github.com/absfs/afero"
)

// An FsActuator makes changes to an afero.Fs.
type FsActuator struct {
	afero.Fs
}

// NewFsActuator returns an actuator that acts on fs.
func NewFsActuator(fs afero.Fs) *FsActuator {
	return &FsActuator{
		Fs: fs,
	}
}

// WriteFile implements Actuator.WriteFile.
func (a *FsActuator) WriteFile(name string, contents []byte, mode os.FileMode, currentContents []byte) error {
	// FIXME use github.com/google/go-write if a.Fs is an afero.OsFs
	return afero.WriteFile(a.Fs, name, contents, mode)
}
