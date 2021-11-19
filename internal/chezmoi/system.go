package chezmoi

import (
	"errors"
	"io/fs"
	"os/exec"
	"sort"

	vfs "github.com/twpayne/go-vfs/v4"
)

// A System reads from and writes to a filesystem, executes idempotent commands,
// runs scripts, and persists state.
type System interface {
	Chmod(name AbsPath, mode fs.FileMode) error
	Glob(pattern string) ([]string, error)
	IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error)
	IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error)
	Link(oldname, newname AbsPath) error
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

func (noUpdateSystemMixin) Chmod(name AbsPath, perm fs.FileMode) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Link(oldname, newname AbsPath) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Mkdir(name AbsPath, perm fs.FileMode) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) RemoveAll(name AbsPath) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Rename(oldpath, newpath AbsPath) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) RunCmd(cmd *exec.Cmd) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) WriteSymlink(oldname string, newname AbsPath) error {
	panic("update to no update system")
}

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

// Walk walks rootAbsPath in system, alling walkFunc for each file or directory in
// the tree, including rootAbsPath.
//
// Walk does not follow symlinks.
func Walk(system System, rootAbsPath AbsPath, walkFunc func(absPath AbsPath, info fs.FileInfo, err error) error) error {
	return vfs.Walk(system.UnderlyingFS(), rootAbsPath.String(), func(absPath string, info fs.FileInfo, err error) error {
		return walkFunc(NewAbsPath(absPath).ToSlash(), info, err)
	})
}

// A WalkSourceDirFunc is a function called for every in a source directory.
type WalkSourceDirFunc func(AbsPath, fs.FileInfo, error) error

// WalkSourceDir walks the source directory rooted at sourceDirAbsPath in
// system, calling walkFunc for each file or directory in the tree, including
// sourceDirAbsPath.
//
// WalkSourceDir does not follow symbolic links found in directories, but if
// sourceDirAbsPath itself is a symbolic link, its target will be walked.
//
// Directory entries .chezmoidata.<format> and .chezmoitemplates are visited
// before all other entries. All other entries are visited in alphabetical
// order.
func WalkSourceDir(system System, sourceDirAbsPath AbsPath, walkFunc WalkSourceDirFunc) error {
	info, err := system.Stat(sourceDirAbsPath)
	if err != nil {
		err = walkFunc(sourceDirAbsPath, nil, err)
	} else {
		err = walkSourceDir(system, sourceDirAbsPath, info, walkFunc)
	}
	if errors.Is(err, fs.SkipDir) {
		return nil
	}
	return err
}

// sourceDirEntryOrder defines the order in which entries are visited in the
// source directory. More negative values are visited first. Entries with the
// same order are visited alphabetically. The default order is zero.
var sourceDirEntryOrder = map[string]int{
	".chezmoidata.json": -2,
	".chezmoidata.toml": -2,
	".chezmoidata.yaml": -2,
	".chezmoitemplates": -1,
}

// walkSourceDir is a helper function for WalkSourceDir.
func walkSourceDir(system System, name AbsPath, info fs.FileInfo, walkFunc WalkSourceDirFunc) error {
	switch err := walkFunc(name, info, nil); {
	case info.IsDir() && errors.Is(err, fs.SkipDir):
		return nil
	case err != nil:
		return err
	case !info.IsDir():
		return nil
	}

	dirEntries, err := system.ReadDir(name)
	if err != nil {
		err = walkFunc(name, info, err)
		if err != nil {
			return err
		}
	}

	sort.Slice(dirEntries, func(i, j int) bool {
		nameI := dirEntries[i].Name()
		nameJ := dirEntries[j].Name()
		orderI := sourceDirEntryOrder[nameI]
		orderJ := sourceDirEntryOrder[nameJ]
		switch {
		case orderI < orderJ:
			return true
		case orderI == orderJ:
			return nameI < nameJ
		default:
			return false
		}
	})

	for _, dirEntry := range dirEntries {
		info, err := dirEntry.Info()
		if err != nil {
			err = walkFunc(name, nil, err)
			if err != nil {
				return err
			}
		}
		if err := walkSourceDir(system, name.JoinString(dirEntry.Name()), info, walkFunc); err != nil {
			if errors.Is(err, fs.SkipDir) {
				break
			}
			return err
		}
	}

	return nil
}
