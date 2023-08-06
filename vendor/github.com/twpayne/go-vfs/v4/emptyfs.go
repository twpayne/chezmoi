package vfs

import (
	"io/fs"
	"os"
	"time"
)

// An EmptyFS is a VFS that does not contain any files.
type EmptyFS struct{}

func (EmptyFS) Chmod(name string, mode fs.FileMode) error         { return os.ErrNotExist }
func (EmptyFS) Chown(name string, uid, git int) error             { return os.ErrNotExist }
func (EmptyFS) Chtimes(name string, atime, mtime time.Time) error { return os.ErrNotExist }
func (EmptyFS) Create(name string) (*os.File, error)              { return nil, os.ErrNotExist }
func (EmptyFS) Glob(pattern string) ([]string, error)             { return nil, os.ErrNotExist }
func (EmptyFS) Lchown(name string, uid, git int) error            { return os.ErrNotExist }
func (EmptyFS) Link(oldname, newname string) error                { return os.ErrNotExist }
func (EmptyFS) Lstat(name string) (fs.FileInfo, error)            { return nil, os.ErrNotExist }
func (EmptyFS) Mkdir(name string, perm fs.FileMode) error         { return os.ErrNotExist }
func (EmptyFS) Open(name string) (fs.File, error)                 { return nil, os.ErrNotExist }
func (EmptyFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	return nil, os.ErrNotExist
}
func (EmptyFS) PathSeparator() rune                                            { return '/' }
func (EmptyFS) RawPath(name string) (string, error)                            { return name, nil }
func (EmptyFS) ReadDir(dirname string) ([]fs.DirEntry, error)                  { return nil, os.ErrNotExist }
func (EmptyFS) ReadFile(filename string) ([]byte, error)                       { return nil, os.ErrNotExist }
func (EmptyFS) Readlink(name string) (string, error)                           { return "", os.ErrNotExist }
func (EmptyFS) Remove(name string) error                                       { return nil }
func (EmptyFS) RemoveAll(name string) error                                    { return nil }
func (EmptyFS) Rename(oldpath, newpath string) error                           { return os.ErrNotExist }
func (EmptyFS) Stat(name string) (fs.FileInfo, error)                          { return nil, os.ErrNotExist }
func (EmptyFS) Symlink(oldname, newname string) error                          { return os.ErrNotExist }
func (EmptyFS) Truncate(name string, size int64) error                         { return os.ErrNotExist }
func (EmptyFS) WriteFile(filename string, data []byte, perm fs.FileMode) error { return os.ErrNotExist }
