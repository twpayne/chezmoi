// +build !windows

package chezmoi

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs"
)

// WriteFile implements Mutator.WriteFile.
func (a *FSMutator) WriteFile(name string, data []byte, perm os.FileMode, currData []byte) error {
	// Special case: if writing to the real filesystem, use github.com/google/renameio
	if a.FS == vfs.OSFS {
		dir := filepath.Dir(name)
		dev, ok := a.devCache[dir]
		if !ok {
			info, err := a.Stat(dir)
			if err != nil {
				return err
			}
			statT, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return errors.New("os.FileInfo.Sys() cannot be converted to a *syscall.Stat_t")
			}
			dev = uint(statT.Dev)
			a.devCache[dir] = dev
		}
		tempDir, ok := a.tempDirCache[dev]
		if !ok {
			tempDir = renameio.TempDir(dir)
			a.tempDirCache[dev] = tempDir
		}
		t, err := renameio.TempFile(tempDir, name)
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
