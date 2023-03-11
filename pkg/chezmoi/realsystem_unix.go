//go:build !windows

package chezmoi

import (
	"errors"
	"io/fs"
	"os"
	"sync"
	"syscall"

	"github.com/google/renameio/v2"
	vfs "github.com/twpayne/go-vfs/v4"
	"go.uber.org/multierr"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fileSystem              vfs.FS
	safe                    bool
	createScriptTempDirOnce sync.Once
	scriptTempDir           AbsPath
	scriptEnv               []string
	devCache                map[AbsPath]uint // devCache maps directories to device numbers.
	tempDirCache            map[uint]string  // tempDirCache maps device numbers to renameio temporary directories.
}

// RealSystemWithSafe sets the safe flag of the RealSystem.
func RealSystemWithSafe(safe bool) RealSystemOption {
	return func(s *RealSystem) {
		s.safe = safe
	}
}

// RealSystemWithScriptTempDir sets the script temporary directory of the RealSystem.
func RealSystemWithScriptTempDir(scriptTempDir AbsPath) RealSystemOption {
	return func(s *RealSystem) {
		s.scriptTempDir = scriptTempDir
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
	return s.fileSystem.Chmod(name.String(), mode)
}

// Readlink implements System.Readlink.
func (s *RealSystem) Readlink(name AbsPath) (string, error) {
	return s.fileSystem.Readlink(name.String())
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) (err error) {
	// Special case: if writing to the real filesystem in safe mode, use
	// github.com/google/renameio.
	if s.safe && s.fileSystem == vfs.OSFS {
		dir := filename.Dir()
		dev, ok := s.devCache[dir]
		if !ok {
			var fileInfo fs.FileInfo
			fileInfo, err = s.Stat(dir)
			if err != nil {
				return err
			}
			statT, ok := fileInfo.Sys().(*syscall.Stat_t)
			if !ok {
				err = errors.New("fs.FileInfo.Sys() cannot be converted to a *syscall.Stat_t")
				return
			}
			dev = uint(statT.Dev)
			s.devCache[dir] = dev
		}
		tempDir, ok := s.tempDirCache[dev]
		if !ok {
			tempDir = renameio.TempDir(dir.String())
			s.tempDirCache[dev] = tempDir
		}
		var t *renameio.PendingFile
		if t, err = renameio.TempFile(tempDir, filename.String()); err != nil {
			return
		}
		defer func() {
			err = multierr.Append(err, t.Cleanup())
		}()
		if err = t.Chmod(perm); err != nil {
			return
		}
		if _, err = t.Write(data); err != nil {
			return
		}
		err = t.CloseAtomicallyReplace()
		return
	}

	return writeFile(s.fileSystem, filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	// Special case: if writing to the real filesystem in safe mode, use
	// github.com/google/renameio.
	if s.safe && s.fileSystem == vfs.OSFS {
		return renameio.Symlink(oldname, newname.String())
	}
	if err := s.fileSystem.RemoveAll(newname.String()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.fileSystem.Symlink(oldname, newname.String())
}

// writeFile is like os.WriteFile but always sets perm before writing data.
// os.WriteFile only sets the permissions when creating a new file. We need to
// ensure permissions, so we use our own implementation.
func writeFile(fileSystem vfs.FS, filename AbsPath, data []byte, perm fs.FileMode) (err error) {
	// Create a new file, or truncate any existing one.
	var f *os.File
	if f, err = fileSystem.OpenFile(filename.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm); err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, f.Close())
	}()

	// Set permissions after truncation but before writing any data, in case the
	// file contained private data before, but before writing the new contents,
	// in case the new contents contain private data after.
	if err = f.Chmod(perm); err != nil {
		return
	}

	_, err = f.Write(data)
	return
}
