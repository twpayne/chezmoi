// Package vfs provides an abstraction of the os and io packages that is easy to
// test.
package vfs

import (
	"io/fs"
	"os"
	"time"
)

// An FS is an abstraction over commonly-used functions in the os and io
// packages.
type FS interface {
	Chmod(name string, mode fs.FileMode) error
	Chown(name string, uid, git int) error
	Chtimes(name string, atime, mtime time.Time) error
	Create(name string) (*os.File, error)
	Glob(pattern string) ([]string, error)
	Lchown(name string, uid, git int) error
	Link(oldname, newname string) error
	Lstat(name string) (fs.FileInfo, error)
	Mkdir(name string, perm fs.FileMode) error
	Open(name string) (fs.File, error)
	OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error)
	PathSeparator() rune
	RawPath(name string) (string, error)
	ReadDir(dirname string) ([]fs.DirEntry, error)
	ReadFile(filename string) ([]byte, error)
	Readlink(name string) (string, error)
	Remove(name string) error
	RemoveAll(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (fs.FileInfo, error)
	Symlink(oldname, newname string) error
	Truncate(name string, size int64) error
	WriteFile(filename string, data []byte, perm fs.FileMode) error
}
