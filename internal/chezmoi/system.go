package chezmoi

import (
	"context"
	"errors"
	"io/fs"
	"os/exec"
	"sort"
	"strings"
	"time"

	vfs "github.com/twpayne/go-vfs/v4"
	"golang.org/x/sync/errgroup"
)

type RunScriptOptions struct {
	Interpreter *Interpreter
	Condition   ScriptCondition
}

// A System reads from and writes to a filesystem, runs scripts, and persists
// state.
type System interface { //nolint:interfacebloat
	Chmod(name AbsPath, mode fs.FileMode) error
	Chtimes(name AbsPath, atime, mtime time.Time) error
	Glob(pattern string) ([]string, error)
	Link(oldname, newname AbsPath) error
	Lstat(filename AbsPath) (fs.FileInfo, error)
	Mkdir(name AbsPath, perm fs.FileMode) error
	RawPath(absPath AbsPath) (AbsPath, error)
	ReadDir(name AbsPath) ([]fs.DirEntry, error)
	ReadFile(name AbsPath) ([]byte, error)
	Readlink(name AbsPath) (string, error)
	Remove(name AbsPath) error
	RemoveAll(name AbsPath) error
	Rename(oldpath, newpath AbsPath) error
	RunCmd(cmd *exec.Cmd) error
	RunScript(scriptname RelPath, dir AbsPath, data []byte, options RunScriptOptions) error
	Stat(name AbsPath) (fs.FileInfo, error)
	UnderlyingFS() vfs.FS
	WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error
	WriteSymlink(oldname string, newname AbsPath) error
}

// A emptySystemMixin simulates an empty system.
type emptySystemMixin struct{}

func (emptySystemMixin) Glob(pattern string) ([]string, error)       { return nil, nil }
func (emptySystemMixin) Lstat(name AbsPath) (fs.FileInfo, error)     { return nil, fs.ErrNotExist }
func (emptySystemMixin) RawPath(path AbsPath) (AbsPath, error)       { return path, nil }
func (emptySystemMixin) ReadDir(name AbsPath) ([]fs.DirEntry, error) { return nil, fs.ErrNotExist }
func (emptySystemMixin) ReadFile(name AbsPath) ([]byte, error)       { return nil, fs.ErrNotExist }
func (emptySystemMixin) Readlink(name AbsPath) (string, error)       { return "", fs.ErrNotExist }
func (emptySystemMixin) Stat(name AbsPath) (fs.FileInfo, error)      { return nil, fs.ErrNotExist }
func (emptySystemMixin) UnderlyingFS() vfs.FS                        { return nil }

// A noUpdateSystemMixin panics on any update.
type noUpdateSystemMixin struct{}

func (noUpdateSystemMixin) Chmod(name AbsPath, perm fs.FileMode) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Chtimes(name AbsPath, atime, mtime time.Time) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Link(oldname, newname AbsPath) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Mkdir(name AbsPath, perm fs.FileMode) error {
	panic("update to no update system")
}

func (noUpdateSystemMixin) Remove(name AbsPath) error {
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

func (noUpdateSystemMixin) RunScript(
	scriptname RelPath,
	dir AbsPath,
	data []byte,
	options RunScriptOptions,
) error {
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
		fileInfo, statErr := system.Stat(absPath)
		if statErr != nil {
			return statErr
		}
		if !fileInfo.IsDir() {
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

// A WalkFunc is called for every entry in a directory.
type WalkFunc func(absPath AbsPath, fileInfo fs.FileInfo, err error) error

// Walk walks rootAbsPath in system, calling walkFunc for each file or directory
// in the tree, including rootAbsPath.
//
// Walk does not follow symlinks.
func Walk(system System, rootAbsPath AbsPath, walkFunc WalkFunc) error {
	outerWalkFunc := func(absPath string, fileInfo fs.FileInfo, err error) error {
		return walkFunc(NewAbsPath(absPath).ToSlash(), fileInfo, err)
	}
	return vfs.Walk(system.UnderlyingFS(), rootAbsPath.String(), outerWalkFunc)
}

// A concurrentWalkSourceDirFunc is a function called concurrently for every
// entry in a source directory.
type concurrentWalkSourceDirFunc func(ctx context.Context, absPath AbsPath, fileInfo fs.FileInfo, err error) error

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
func WalkSourceDir(system System, sourceDirAbsPath AbsPath, walkFunc WalkFunc) error {
	fileInfo, err := system.Stat(sourceDirAbsPath)
	if err != nil {
		err = walkFunc(sourceDirAbsPath, nil, err)
	} else {
		err = walkSourceDir(system, sourceDirAbsPath, fileInfo, walkFunc)
		if errors.Is(err, fs.SkipDir) {
			err = nil
		}
	}
	return err
}

// sourceDirEntryOrder defines the order in which entries are visited in the
// source directory. More negative values are visited first. Entries with the
// same order are visited alphabetically. The default order is zero.
var sourceDirEntryOrder = map[string]int{
	VersionName:        -3,
	dataName + ".json": -2,
	dataName + ".toml": -2,
	dataName + ".yaml": -2,
	TemplatesDirName:   -1,
}

// walkSourceDir is a helper function for WalkSourceDir.
func walkSourceDir(system System, name AbsPath, fileInfo fs.FileInfo, walkFunc WalkFunc) error {
	switch err := walkFunc(name, fileInfo, nil); {
	case fileInfo.IsDir() && errors.Is(err, fs.SkipDir):
		return nil
	case err != nil:
		return err
	case !fileInfo.IsDir():
		return nil
	}

	dirEntries, err := system.ReadDir(name)
	if err != nil {
		err = walkFunc(name, fileInfo, err)
		if err != nil {
			return err
		}
	}

	sortSourceDirEntries(dirEntries)

	for _, dirEntry := range dirEntries {
		fileInfo, err := dirEntry.Info()
		if err != nil {
			err = walkFunc(name, nil, err)
			if err != nil {
				return err
			}
		}
		if err := walkSourceDir(system, name.JoinString(dirEntry.Name()), fileInfo, walkFunc); err != nil {
			if !errors.Is(err, fs.SkipDir) {
				return err
			}
		}
	}

	return nil
}

func concurrentWalkSourceDir(
	ctx context.Context, system System, dirAbsPath AbsPath, walkFunc concurrentWalkSourceDirFunc,
) error {
	dirEntries, err := system.ReadDir(dirAbsPath)
	if err != nil {
		return walkFunc(ctx, dirAbsPath, nil, err)
	}
	sortSourceDirEntries(dirEntries)

	// Walk all control plane entries in order.
	visitDirEntry := func(dirEntry fs.DirEntry) error {
		absPath := dirAbsPath.Join(NewRelPath(dirEntry.Name()))
		fileInfo, err := dirEntry.Info()
		if err != nil {
			return walkFunc(ctx, absPath, nil, err)
		}
		switch err := walkFunc(ctx, absPath, fileInfo, nil); {
		case fileInfo.IsDir() && errors.Is(err, fs.SkipDir):
			return nil
		case err != nil:
			return err
		case fileInfo.IsDir():
			return concurrentWalkSourceDir(ctx, system, absPath, walkFunc)
		default:
			return nil
		}
	}
	i := 0
	for ; i < len(dirEntries); i++ {
		dirEntry := dirEntries[i]
		if !strings.HasPrefix(dirEntry.Name(), ".") {
			break
		}
		if err := visitDirEntry(dirEntry); err != nil {
			return err
		}
	}

	// Walk all remaining entries concurrently.
	visitDirEntryFunc := func(dirEntry fs.DirEntry) func() error {
		return func() error {
			return visitDirEntry(dirEntry)
		}
	}
	group, ctx := errgroup.WithContext(ctx)
	for _, dirEntry := range dirEntries[i:] {
		group.Go(visitDirEntryFunc(dirEntry))
	}
	return group.Wait()
}

func sortSourceDirEntries(dirEntries []fs.DirEntry) {
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
}
