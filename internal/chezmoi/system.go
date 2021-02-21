package chezmoi

import (
	"os"
	"os/exec"
	"path/filepath"

	vfs "github.com/twpayne/go-vfs/v2"
)

// A System reads from and writes to a filesystem, executes idempotent commands,
// runs scripts, and persists state.
type System interface {
	Chmod(name AbsPath, mode os.FileMode) error
	Glob(pattern string) ([]string, error)
	IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error)
	Lstat(filename AbsPath) (os.FileInfo, error)
	Mkdir(name AbsPath, perm os.FileMode) error
	RawPath(absPath AbsPath) (AbsPath, error)
	ReadDir(name AbsPath) ([]os.DirEntry, error)
	ReadFile(name AbsPath) ([]byte, error)
	Readlink(name AbsPath) (string, error)
	RemoveAll(name AbsPath) error
	Rename(oldpath, newpath AbsPath) error
	RunCmd(cmd *exec.Cmd) error
	RunScript(scriptname RelPath, dir AbsPath, data []byte) error
	Stat(name AbsPath) (os.FileInfo, error)
	UnderlyingFS() vfs.FS
	WriteFile(filename AbsPath, data []byte, perm os.FileMode) error
	WriteSymlink(oldname string, newname AbsPath) error
}

// A emptySystemMixin simulates an empty system.
type emptySystemMixin struct{}

func (emptySystemMixin) Glob(pattern string) ([]string, error)             { return nil, nil }
func (emptySystemMixin) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) { return nil, nil }
func (emptySystemMixin) Lstat(name AbsPath) (os.FileInfo, error)           { return nil, os.ErrNotExist }
func (emptySystemMixin) RawPath(path AbsPath) (AbsPath, error)             { return path, nil }
func (emptySystemMixin) ReadDir(name AbsPath) ([]os.DirEntry, error)       { return nil, os.ErrNotExist }
func (emptySystemMixin) ReadFile(name AbsPath) ([]byte, error)             { return nil, os.ErrNotExist }
func (emptySystemMixin) Readlink(name AbsPath) (string, error)             { return "", os.ErrNotExist }
func (emptySystemMixin) Stat(name AbsPath) (os.FileInfo, error)            { return nil, os.ErrNotExist }
func (emptySystemMixin) UnderlyingFS() vfs.FS                              { return nil }

// A noUpdateSystemMixin panics on any update.
type noUpdateSystemMixin struct{}

func (noUpdateSystemMixin) Chmod(name AbsPath, perm os.FileMode) error                   { panic(nil) }
func (noUpdateSystemMixin) Mkdir(name AbsPath, perm os.FileMode) error                   { panic(nil) }
func (noUpdateSystemMixin) RemoveAll(name AbsPath) error                                 { panic(nil) }
func (noUpdateSystemMixin) Rename(oldpath, newpath AbsPath) error                        { panic(nil) }
func (noUpdateSystemMixin) RunCmd(cmd *exec.Cmd) error                                   { panic(nil) }
func (noUpdateSystemMixin) RunScript(scriptname RelPath, dir AbsPath, data []byte) error { panic(nil) }
func (noUpdateSystemMixin) WriteFile(filename AbsPath, data []byte, perm os.FileMode) error {
	panic(nil)
}
func (noUpdateSystemMixin) WriteSymlink(oldname string, newname AbsPath) error { panic(nil) }

// MkdirAll is the equivalent of os.MkdirAll but operates on system.
func MkdirAll(system System, absPath AbsPath, perm os.FileMode) error {
	switch err := system.Mkdir(absPath, perm); {
	case err == nil:
		// Mkdir was successful.
		return nil
	case os.IsExist(err):
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
	case os.IsNotExist(err):
		// Parent directory does not exist. Create the parent directory
		// recursively, then try again.
		parentDir := absPath.Dir()
		if parentDir == "/" || parentDir == "." {
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

// Walk walks rootAbsPath in s.
func Walk(system System, rootAbsPath AbsPath, walkFn func(absPath AbsPath, info os.FileInfo, err error) error) error {
	return vfs.Walk(system.UnderlyingFS(), string(rootAbsPath), func(absPath string, info os.FileInfo, err error) error {
		return walkFn(AbsPath(filepath.ToSlash(absPath)), info, err)
	})
}
