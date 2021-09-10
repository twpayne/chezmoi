//go:build !windows
// +build !windows

package chezmoi

import (
	"errors"
	"io/fs"
	"os"
	"syscall"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs/v4"
	"go.uber.org/multierr"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fileSystem   vfs.FS
	safe         bool
	devCache     map[AbsPath]uint // devCache maps directories to device numbers.
	tempDirCache map[uint]string  // tempDirCache maps device numbers to renameio temporary directories.
}

// RealSystemWithSafe sets the safe flag of the RealSystem.
func RealSystemWithSafe(safe bool) RealSystemOption {
	return func(s *RealSystem) {
		s.safe = safe
	}
}

// NewRealSystem returns a System that acts on fileSystem.
func NewRealSystem(fileSystem vfs.FS, options ...RealSystemOption) *RealSystem {
	s := &RealSystem{
		fileSystem:   fileSystem,
		safe:         true,
		devCache:     make(map[AbsPath]uint),
		tempDirCache: make(map[uint]string),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// Chmod implements System.Chmod.
func (s *RealSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	return s.fileSystem.Chmod(string(name), mode)
}

// Readlink implements System.Readlink.
func (s *RealSystem) Readlink(name AbsPath) (string, error) {
	return s.fileSystem.Readlink(string(name))
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	// Special case: if writing to the real filesystem in safe mode, use
	// github.com/google/renameio.
	if s.safe && s.fileSystem == vfs.OSFS {
		dir := filename.Dir()
		dev, ok := s.devCache[dir]
		if !ok {
			info, err := s.Stat(dir)
			if err != nil {
				return err
			}
			statT, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return errors.New("fs.FileInfo.Sys() cannot be converted to a *syscall.Stat_t")
			}
			dev = uint(statT.Dev)
			s.devCache[dir] = dev
		}
		tempDir, ok := s.tempDirCache[dev]
		if !ok {
			tempDir = renameio.TempDir(string(dir))
			s.tempDirCache[dev] = tempDir
		}
		t, err := renameio.TempFile(tempDir, string(filename))
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

	return writeFile(s.fileSystem, filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	// Special case: if writing to the real filesystem in safe mode, use
	// github.com/google/renameio.
	if s.safe && s.fileSystem == vfs.OSFS {
		return renameio.Symlink(oldname, string(newname))
	}
	if err := s.fileSystem.RemoveAll(string(newname)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.fileSystem.Symlink(oldname, string(newname))
}

// writeFile is like os.WriteFile but always sets perm before writing data.
// os.WriteFile only sets the permissions when creating a new file. We need to
// ensure permissions, so we use our own implementation.
func writeFile(fileSystem vfs.FS, filename AbsPath, data []byte, perm fs.FileMode) (err error) {
	// Create a new file, or truncate any existing one.
	f, err := fileSystem.OpenFile(string(filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, f.Close())
	}()

	// Set permissions after truncation but before writing any data, in case the
	// file contained private data before, but before writing the new contents,
	// in case the contents contain private data after.
	if err = f.Chmod(perm); err != nil {
		return
	}

	_, err = f.Write(data)
	return
}
