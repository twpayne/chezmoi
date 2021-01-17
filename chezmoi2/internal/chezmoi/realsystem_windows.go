package chezmoi

import (
	"os"
	"path/filepath"

	vfs "github.com/twpayne/go-vfs"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fs vfs.FS
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fs vfs.FS) *RealSystem {
	return &RealSystem{
		fs: fs,
	}
}

// Chmod implements System.Chmod.
func (s *RealSystem) Chmod(name AbsPath, mode os.FileMode) error {
	return nil
}

// Readlink implements System.Readlink.
func (s *RealSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.fs.Readlink(string(name))
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(linkname), nil
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm os.FileMode) error {
	return s.fs.WriteFile(string(filename), data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	if err := s.fs.RemoveAll(string(newname)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.fs.Symlink(filepath.FromSlash(oldname), string(newname))
}
