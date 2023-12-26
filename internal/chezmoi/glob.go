package chezmoi

import (
	"io/fs"

	"github.com/bmatcuk/doublestar/v4"
	vfs "github.com/twpayne/go-vfs/v4"
)

// A lstatFS implements io/fs.StatFS but uses Lstat instead of Stat.
type lstatFS struct {
	wrapped interface {
		fs.FS
		Lstat(name string) (fs.FileInfo, error)
	}
}

// Open implements io/fs.StatFS.Open.
func (s lstatFS) Open(name string) (fs.File, error) {
	return s.wrapped.Open(name)
}

// Stat implements io/fs.StatFS.Stat.
func (s lstatFS) Stat(name string) (fs.FileInfo, error) {
	return s.wrapped.Lstat(name)
}

// Glob is like github.com/bmatcuk/doublestar/v4.Glob except that it does not
// follow symlinks.
func Glob(fileSystem vfs.FS, prefix string) ([]string, error) {
	fsys := lstatFS{wrapped: fileSystem}
	opts := []doublestar.GlobOption{
		doublestar.WithFailOnIOErrors(),
		doublestar.WithNoFollow(),
	}
	return doublestar.Glob(fsys, prefix, opts...)
}
