package chezmoi

import (
	"os"
	"os/exec"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
)

// An FSMutator makes changes to a vfs.FS.
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

// IdempotentCmdOutput implements Mutator.IdempotentCmdOutput.
func (m *FSMutator) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

// RunCmd implements Mutator.RunCmd.
func (m *FSMutator) RunCmd(cmd *exec.Cmd) error {
	return cmd.Run()
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *FSMutator) WriteSymlink(oldname, newname string) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if m.FS == vfs.OSFS {
		return renameio.Symlink(oldname, newname)
	}
	if err := m.FS.RemoveAll(newname); err != nil && !os.IsNotExist(err) {
		return err
	}
	return m.FS.Symlink(oldname, newname)
}
