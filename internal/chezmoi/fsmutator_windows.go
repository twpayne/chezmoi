// +build windows

package chezmoi

import (
	"os"
)

// WriteFile implements Mutator.WriteFile.
func (m *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	return m.FS.WriteFile(name, data, perm)
}

// WriteSymlink implements Mutator.WriteSymlink.
func (m *FSMutator) WriteSymlink(oldname, newname string) error {
	if err := m.FS.RemoveAll(newname); err != nil && !os.IsNotExist(err) {
		return err
	}
	return m.FS.Symlink(oldname, newname)
}
