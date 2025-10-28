package chezmoi

import (
	"io/fs"

	"github.com/bmatcuk/doublestar/v4"
	vfs "github.com/twpayne/go-vfs/v5"
)

// A LstatFS implements io/fs.StatFS but uses Lstat instead of Stat.
type LstatFS struct {
	Wrapped interface {
		fs.FS
		Lstat(name string) (fs.FileInfo, error)
	}
}

// Open implements io/fs.StatFS.Open.
func (s LstatFS) Open(name string) (fs.File, error) {
	return s.Wrapped.Open(name)
}

// Stat implements io/fs.StatFS.Stat.
func (s LstatFS) Stat(name string) (fs.FileInfo, error) {
	return s.Wrapped.Lstat(name)
}

// Glob is like github.com/bmatcuk/doublestar/v4.Glob except that it does not
// follow symlinks.
func Glob(fileSystem vfs.FS, prefix string) ([]string, error) {
	fsys := LstatFS{Wrapped: fileSystem}
	opts := []doublestar.GlobOption{
		doublestar.WithFailOnIOErrors(),
		doublestar.WithNoFollow(),
	}
	return doublestar.Glob(fsys, prefix, opts...)
}
