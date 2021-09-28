package chezmoi

import (
	"errors"
	"io/fs"
	"path/filepath"

	vfs "github.com/twpayne/go-vfs/v4"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fileSystem vfs.FS
}

// RealSystemWithSafe sets the safe flag of the RealSystem. On Windows it does
// nothing as Windows does not support atomic file or symlink updates. See
// https://github.com/google/renameio/issues/1 and
// https://github.com/golang/go/issues/22397#issuecomment-498856679.
func RealSystemWithSafe(safe bool) RealSystemOption {
	return func(s *RealSystem) {}
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fileSystem vfs.FS, options ...RealSystemOption) *RealSystem {
	return &RealSystem{
		fileSystem: fileSystem,
	}
}

// Chmod implements System.Chmod.
func (s *RealSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	return nil
}

// Readlink implements System.Readlink.
func (s *RealSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.fileSystem.Readlink(name.String())
	if err != nil {
		return "", err
	}
	return normalizeLinkname(linkname), nil
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	return s.fileSystem.WriteFile(filename.String(), data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	if err := s.fileSystem.RemoveAll(newname.String()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.fileSystem.Symlink(filepath.FromSlash(oldname), newname.String())
}
