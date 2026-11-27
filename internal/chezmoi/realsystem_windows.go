package chezmoi

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sync"

	vfs "github.com/twpayne/go-vfs/v5"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fileSystem              vfs.FS
	createScriptTempDirOnce sync.Once
	scriptTempDir           AbsPath
}

// RealSystemWithSafe sets the safe flag of the RealSystem. On Windows it does
// nothing as Windows does not support atomic file or symlink updates. See
// https://github.com/google/renameio/issues/1 and
// https://github.com/golang/go/issues/22397#issuecomment-498856679.
func RealSystemWithSafe(safe bool) RealSystemOption {
	return func(s *RealSystem) {}
}

// RealSystemWithScriptTempDir sets the script temporary directory of the
// RealSystem.
func RealSystemWithScriptTempDir(scriptTempDir AbsPath) RealSystemOption {
	return func(s *RealSystem) {}
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fileSystem vfs.FS, options ...RealSystemOption) *RealSystem {
	s := &RealSystem{
		fileSystem: fileSystem,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// Chmod implements System.Chmod.
func (s *RealSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	return s.fileSystem.Chmod(name.String(), mode)
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
	// First, try to write the file. If the error is anything except a
	// permission error, return that error.
	writeFileErr := s.fileSystem.WriteFile(filename.String(), data, perm)
	if !errors.Is(writeFileErr, fs.ErrPermission) {
		return writeFileErr
	}

	// If we got a permission error, it might be because the file is read only.
	// Lstat the file to find out. If the lstat fails, return that error.
	fileInfo, lstatErr := s.fileSystem.Lstat(filename.String())
	if lstatErr != nil {
		return lstatErr
	}

	// If the file is not read-only, then return the original error.
	if fileInfo.Mode().Perm()&0o222 != 0 {
		return writeFileErr
	}

	// Chmod the file to enable write permissions.
	if chmodErr := s.fileSystem.Chmod(filename.String(), 0o600); chmodErr != nil {
		return chmodErr
	}

	// Ensure that we try to leave the file with the correct permissions. Ignore
	// any errors because there's nothing else we can do.
	defer func() {
		_ = s.fileSystem.Chmod(filename.String(), perm)
	}()

	// Write the file and rely on the previous defer to re-establish the
	// read-only permissions, even if the write fails.
	return s.fileSystem.WriteFile(filename.String(), data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldName string, newName AbsPath) error {
	if err := s.fileSystem.RemoveAll(newName.String()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.fileSystem.Symlink(filepath.FromSlash(oldName), newName.String())
}
