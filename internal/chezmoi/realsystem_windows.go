package chezmoi

import (
	"errors"
	"io/fs"
	"path/filepath"

	vfs "github.com/twpayne/go-vfs/v3"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fileSystem vfs.FS
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fileSystem vfs.FS) *RealSystem {
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
	linkname, err := s.fileSystem.Readlink(string(name))
	if err != nil {
		return "", err
	}
	return normalizeLinkname(linkname), nil
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	return s.fileSystem.WriteFile(string(filename), data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	if err := s.fileSystem.RemoveAll(string(newname)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.fileSystem.Symlink(filepath.FromSlash(oldname), string(newname))
}
