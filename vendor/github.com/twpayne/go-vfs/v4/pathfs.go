package vfs

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

// A PathFS operates on an existing FS, but prefixes all names with a path. All
// names must be absolute paths, with the exception of symlinks, which may be
// relative.
type PathFS struct {
	fileSystem FS
	path       string
}

// NewPathFS returns a new *PathFS operating on fileSystem and prefixing all
// names with path.
func NewPathFS(fileSystem FS, path string) *PathFS {
	return &PathFS{
		fileSystem: fileSystem,
		path:       filepath.ToSlash(path),
	}
}

// Chmod implements os.Chmod.
func (p *PathFS) Chmod(name string, mode fs.FileMode) error {
	realName, err := p.join("Chmod", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Chmod(realName, mode)
}

// Chown implements os.Chown.
func (p *PathFS) Chown(name string, uid, gid int) error {
	realName, err := p.join("Chown", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Chown(realName, uid, gid)
}

// Chtimes implements os.Chtimes.
func (p *PathFS) Chtimes(name string, atime, mtime time.Time) error {
	realName, err := p.join("Chtimes", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Chtimes(realName, atime, mtime)
}

// Create implements os.Create.
func (p *PathFS) Create(name string) (*os.File, error) {
	realName, err := p.join("Create", name)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.Create(realName)
}

// Glob implements filepath.Glob.
func (p *PathFS) Glob(pattern string) ([]string, error) {
	realPattern, err := p.join("Glob", pattern)
	if err != nil {
		return nil, err
	}
	matches, err := p.fileSystem.Glob(realPattern)
	if err != nil {
		return nil, err
	}
	for i, match := range matches {
		matches[i], err = trimPrefix(match, p.path)
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}

// Join returns p's path joined with name.
func (p *PathFS) Join(op, name string) (string, error) {
	return p.join("Join", name)
}

// Lchown implements os.Lchown.
func (p *PathFS) Lchown(name string, uid, gid int) error {
	realName, err := p.join("Lchown", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Lchown(realName, uid, gid)
}

// Link implements os.Link.
func (p *PathFS) Link(oldname, newname string) error {
	var realOldname string
	if path.IsAbs(oldname) {
		var err error
		realOldname, err = p.join("Link", oldname)
		if err != nil {
			return err
		}
	} else {
		realOldname = oldname
	}
	realNewname, err := p.join("Link", newname)
	if err != nil {
		return err
	}
	return p.fileSystem.Link(realOldname, realNewname)
}

// Lstat implements os.Lstat.
func (p *PathFS) Lstat(name string) (fs.FileInfo, error) {
	realName, err := p.join("Lstat", name)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.Lstat(realName)
}

// Mkdir implements os.Mkdir.
func (p *PathFS) Mkdir(name string, perm fs.FileMode) error {
	realName, err := p.join("Mkdir", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Mkdir(realName, perm)
}

// Open implements os.Open.
func (p *PathFS) Open(name string) (fs.File, error) {
	realName, err := p.join("Open", name)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.Open(realName)
}

// OpenFile implements os.OpenFile.
func (p *PathFS) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	realName, err := p.join("OpenFile", name)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.OpenFile(realName, flag, perm)
}

// PathSeparator implements PathSeparator.
func (p *PathFS) PathSeparator() rune {
	return p.fileSystem.PathSeparator()
}

// RawPath implements RawPath.
func (p *PathFS) RawPath(path string) (string, error) {
	return p.join("RawPath", path)
}

// ReadDir implements os.ReadDir.
func (p *PathFS) ReadDir(dirname string) ([]fs.DirEntry, error) {
	realDirname, err := p.join("ReadDir", dirname)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.ReadDir(realDirname)
}

// ReadFile implements os.ReadFile.
func (p *PathFS) ReadFile(name string) ([]byte, error) {
	realName, err := p.join("ReadFile", name)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.ReadFile(realName)
}

// Readlink implements os.Readlink.
func (p *PathFS) Readlink(name string) (string, error) {
	realName, err := p.join("Readlink", name)
	if err != nil {
		return "", err
	}
	return p.fileSystem.Readlink(realName)
}

// Remove implements os.Remove.
func (p *PathFS) Remove(name string) error {
	realName, err := p.join("Remove", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Remove(realName)
}

// RemoveAll implements os.RemoveAll.
func (p *PathFS) RemoveAll(name string) error {
	realName, err := p.join("RemoveAll", name)
	if err != nil {
		return err
	}
	return p.fileSystem.RemoveAll(realName)
}

// Rename implements os.Rename.
func (p *PathFS) Rename(oldpath, newpath string) error {
	realOldpath, err := p.join("Rename", oldpath)
	if err != nil {
		return err
	}
	realNewpath, err := p.join("Rename", newpath)
	if err != nil {
		return err
	}
	return p.fileSystem.Rename(realOldpath, realNewpath)
}

// Stat implements os.Stat.
func (p *PathFS) Stat(name string) (fs.FileInfo, error) {
	realName, err := p.join("Stat", name)
	if err != nil {
		return nil, err
	}
	return p.fileSystem.Stat(realName)
}

// Symlink implements os.Symlink.
func (p *PathFS) Symlink(oldname, newname string) error {
	var realOldname string
	if path.IsAbs(oldname) {
		var err error
		realOldname, err = p.join("Symlink", oldname)
		if err != nil {
			return err
		}
	} else {
		realOldname = oldname
	}
	realNewname, err := p.join("Symlink", newname)
	if err != nil {
		return err
	}
	return p.fileSystem.Symlink(realOldname, realNewname)
}

// Truncate implements os.Truncate.
func (p *PathFS) Truncate(name string, size int64) error {
	realName, err := p.join("Truncate", name)
	if err != nil {
		return err
	}
	return p.fileSystem.Truncate(realName, size)
}

// WriteFile implements io.WriteFile.
func (p *PathFS) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	realFilename, err := p.join("WriteFile", filename)
	if err != nil {
		return err
	}
	return p.fileSystem.WriteFile(realFilename, data, perm)
}

// join returns p's path joined with name.
func (p *PathFS) join(op, name string) (string, error) {
	name = relativizePath(name)
	if !path.IsAbs(name) {
		return "", &os.PathError{
			Op:   op,
			Path: name,
			Err:  syscall.EPERM,
		}
	}
	return filepath.Join(p.path, name), nil
}
