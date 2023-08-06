package vfs

import (
	"errors"
	"io/fs"
	"path/filepath"
)

// A MkdirStater implements all the functionality needed by MkdirAll.
type MkdirStater interface {
	Mkdir(name string, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
}

// MkdirAll is equivalent to os.MkdirAll but operates on fileSystem.
func MkdirAll(fileSystem MkdirStater, path string, perm fs.FileMode) error {
	err := fileSystem.Mkdir(path, perm)
	switch {
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
		info, statErr := fileSystem.Stat(path)
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
		parentDir := filepath.Dir(path)
		if parentDir == "/" || parentDir == "." {
			// We cannot create the root directory or the current directory, so
			// return the original error.
			return err
		}
		if err := MkdirAll(fileSystem, parentDir, perm); err != nil {
			return err
		}
		return fileSystem.Mkdir(path, perm)
	default:
		// Some other error.
		return err
	}
}
