// +build windows

package chezmoi

import (
	"os"
)

// WriteFile implements Mutator.WriteFile.
func (m *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	return m.FS.WriteFile(name, data, perm)
}
