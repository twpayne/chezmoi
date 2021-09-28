package chezmoi

import (
	"errors"
	"io/fs"
	"os/exec"

	vfs "github.com/twpayne/go-vfs/v4"
)

// A System reads from and writes to a filesystem, executes idempotent commands,
// runs scripts, and persists state.
type System interface {
	Chmod(name AbsPath, mode fs.FileMode) error
	Glob(pattern string) ([]string, error)
	IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error)
	IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error)
	Lstat(filename AbsPath) (fs.FileInfo, error)
	Mkdir(name AbsPath, perm fs.FileMode) error
	RawPath(absPath AbsPath) (AbsPath, error)
	ReadDir(name AbsPath) ([]fs.DirEntry, error)
	ReadFile(name AbsPath) ([]byte, error)
	Readlink(name AbsPath) (string, error)
	RemoveAll(name AbsPath) error
	Rename(oldpath, newpath AbsPath) error
	RunCmd(cmd *exec.Cmd) error
	RunIdempotentCmd(cmd *exec.Cmd) error
	RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error
	Stat(name AbsPath) (fs.FileInfo, error)
	UnderlyingFS() vfs.FS
	WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error
	WriteSymlink(oldname string, newname AbsPath) error
}

// A emptySystemMixin simulates an empty system.
type emptySystemMixin struct{}

func (emptySystemMixin) Glob(pattern string) ([]string, error)                     { return nil, nil }
func (emptySystemMixin) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) { return nil, nil }
func (emptySystemMixin) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error)         { return nil, nil }
func (emptySystemMixin) Lstat(name AbsPath) (fs.FileInfo, error)                   { return nil, fs.ErrNotExist }
func (emptySystemMixin) RawPath(path AbsPath) (AbsPath, error)                     { return path, nil }
func (emptySystemMixin) ReadDir(name AbsPath) ([]fs.DirEntry, error)               { return nil, fs.ErrNotExist }
func (emptySystemMixin) ReadFile(name AbsPath) ([]byte, error)                     { return nil, fs.ErrNotExist }
func (emptySystemMixin) Readlink(name AbsPath) (string, error)                     { return "", fs.ErrNotExist }
func (emptySystemMixin) RunIdempotentCmd(cmd *exec.Cmd) error                      { return nil }
func (emptySystemMixin) Stat(name AbsPath) (fs.FileInfo, error)                    { return nil, fs.ErrNotExist }
func (emptySystemMixin) UnderlyingFS() vfs.FS                                      { return nil }

// A noUpdateSystemMixin panics on any update.
type noUpdateSystemMixin struct{}

func (noUpdateSystemMixin) Chmod(name AbsPath, perm fs.FileMode) error { panic(nil) }
func (noUpdateSystemMixin) Mkdir(name AbsPath, perm fs.FileMode) error { panic(nil) }
func (noUpdateSystemMixin) RemoveAll(name AbsPath) error               { panic(nil) }
func (noUpdateSystemMixin) Rename(oldpath, newpath AbsPath) error      { panic(nil) }
func (noUpdateSystemMixin) RunCmd(cmd *exec.Cmd) error                 { panic(nil) }
func (noUpdateSystemMixin) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	panic(nil)
}

func (noUpdateSystemMixin) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	panic(nil)
}
func (noUpdateSystemMixin) WriteSymlink(oldname string, newname AbsPath) error { panic(nil) }

// MkdirAll is the equivalent of os.MkdirAll but operates on system.
func MkdirAll(system System, absPath AbsPath, perm fs.FileMode) error {
	switch err := system.Mkdir(absPath, perm); {
	case err == nil:
		// Mkdir was successful.
		return nil
	case errors.Is(err, fs.ErrExist):
		// path already exists, but we don't know whether it's a directory or
		// something else. We get this error if we try to create a subdirectory
		// of a non-directory, for example if the parent directory of path is a
		// file. There's a race condition here between the call to Mkdir and the
		// call to Stat but we can't avoid it because there's not enough
		// information in the returned error from Mkdir. We need to distinguish
		// between "path already exists and is already a directory" and "path
		// already exists and is not a directory". Between the call to Mkdir and
		// the call to Stat path might have changed.
		info, statErr := system.Stat(absPath)
		if statErr != nil {
			return statErr
		}
		if !info.IsDir() {
			return err
		}
		return nil
	case errors.Is(err, fs.ErrNotExist):
		// Parent directory does not exist. Create the parent directory
		// recursively, then try again.
		parentDir := absPath.Dir()
		if parentDir == RootAbsPath || parentDir == DotAbsPath {
			// We cannot create the root directory or the current directory, so
			// return the original error.
			return err
		}
		if err := MkdirAll(system, parentDir, perm); err != nil {
			return err
		}
		return system.Mkdir(absPath, perm)
	default:
		// Some other error.
		return err
	}
}

// Walk walks rootAbsPath in system, alling walkFn for each file or directory in
// the tree, including rootAbsPath.
//
// Walk does not follow symlinks.
func Walk(system System, rootAbsPath AbsPath, walkFn func(absPath AbsPath, info fs.FileInfo, err error) error) error {
	return vfs.Walk(system.UnderlyingFS(), rootAbsPath.String(), func(absPath string, info fs.FileInfo, err error) error {
		return walkFn(NewAbsPath(absPath).ToSlash(), info, err)
	})
}

// WalkDir walks the file tree rooted at rootAbsPath in system, calling walkFn
// for each file or directory in the tree, including rootAbsPath.
//
// WalkDir does not follow symbolic links found in directories, but if
// rootAbsPath itself is a symbolic link, its target will be walked.
func WalkDir(system System, rootAbsPath AbsPath, walkFn func(absPath AbsPath, info fs.FileInfo, err error) error) error {
	return fs.WalkDir(system.UnderlyingFS(), rootAbsPath.String(), func(path string, dirEntry fs.DirEntry, err error) error {
		var info fs.FileInfo
		if err == nil {
			info, err = dirEntry.Info()
		}
		return walkFn(NewAbsPath(path).ToSlash(), info, err)
	})
}
