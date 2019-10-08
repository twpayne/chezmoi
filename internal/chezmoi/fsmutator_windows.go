// +build windows

package chezmoi

import (
	"os"
)

// WriteFile implements Mutator.WriteFile.
func (a *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	return a.FS.WriteFile(name, data, perm)
}
