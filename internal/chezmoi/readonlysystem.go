package chezmoi

import (
	"io/fs"

	vfs "github.com/twpayne/go-vfs/v5"
)

// A ReadOnlySystem is a system that may only be read from.
type ReadOnlySystem struct {
	noUpdateSystemMixin

	system System
}

// NewReadOnlySystem returns a new ReadOnlySystem that wraps system.
func NewReadOnlySystem(system System) *ReadOnlySystem {
	return &ReadOnlySystem{
		system: system,
	}
}

// Glob implements System.Glob.
func (s *ReadOnlySystem) Glob(pattern string) ([]string, error) {
	return s.system.Glob(pattern)
}

// Lstat implements System.Lstat.
func (s *ReadOnlySystem) Lstat(filename AbsPath) (fs.FileInfo, error) {
	return s.system.Lstat(filename)
}

// RawPath implements System.RawPath.
func (s *ReadOnlySystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *ReadOnlySystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.system.ReadDir(name)
}

// ReadFile implements System.ReadFile.
func (s *ReadOnlySystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.system.ReadFile(name)
}

// Readlink implements System.Readlink.
func (s *ReadOnlySystem) Readlink(name AbsPath) (string, error) {
	return s.system.Readlink(name)
}

// Stat implements System.Stat.
func (s *ReadOnlySystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Stat(name)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *ReadOnlySystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}
