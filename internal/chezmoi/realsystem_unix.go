// +build !windows

package chezmoi

import (
	"errors"
	"os"
	"syscall"

	"github.com/google/renameio"
	vfs "github.com/twpayne/go-vfs/v2"
	"go.uber.org/multierr"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fs           vfs.FS
	devCache     map[AbsPath]uint // devCache maps directories to device numbers.
	tempDirCache map[uint]string  // tempDirCache maps device numbers to renameio temporary directories.
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fs vfs.FS) *RealSystem {
	return &RealSystem{
		fs:           fs,
		devCache:     make(map[AbsPath]uint),
		tempDirCache: make(map[uint]string),
	}
}

// Chmod implements System.Chmod.
func (s *RealSystem) Chmod(name AbsPath, mode os.FileMode) error {
	return s.fs.Chmod(string(name), mode)
}

// Readlink implements System.Readlink.
func (s *RealSystem) Readlink(name AbsPath) (string, error) {
	return s.fs.Readlink(string(name))
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm os.FileMode) error {
	// Special case: if writing to the real filesystem, use
	// github.com/google/renameio.
	if s.fs == vfs.OSFS {
		dir := filename.Dir()
		dev, ok := s.devCache[dir]
		if !ok {
			info, err := s.Stat(dir)
			if err != nil {
				return err
			}
			statT, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return errors.New("os.FileInfo.Sys() cannot be converted to a *syscall.Stat_t")
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

	return writeFile(s.fs, filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	// Special case: if writing to the real filesystem, use
	// github.com/google/renameio.
	if s.fs == vfs.OSFS {
		return renameio.Symlink(oldname, string(newname))
	}
	if err := s.fs.RemoveAll(string(newname)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return s.fs.Symlink(oldname, string(newname))
}

// writeFile is like os.WriteFile but always sets perm before writing data.
// os.WriteFile only sets the permissions when creating a new file. We need to
// ensure permissions, so we use our own implementation.
func writeFile(fs vfs.FS, filename AbsPath, data []byte, perm os.FileMode) (err error) {
	// Create a new file, or truncate any existing one.
	f, err := fs.OpenFile(string(filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
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
