package vfs

//nolint:godox
// FIXME implement path/filepath.WalkDir

import (
	"errors"
	"io/fs"
	"path/filepath"
	"sort"
)

// SkipDir is fs.SkipDir.
var SkipDir = fs.SkipDir

// A LstatReadDirer implements all the functionality needed by Walk.
type LstatReadDirer interface {
	Lstat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}

type dirEntriesByName []fs.DirEntry

func (is dirEntriesByName) Len() int           { return len(is) }
func (is dirEntriesByName) Less(i, j int) bool { return is[i].Name() < is[j].Name() }
func (is dirEntriesByName) Swap(i, j int)      { is[i], is[j] = is[j], is[i] }

// walk recursively walks fileSystem from path.
func walk(fileSystem LstatReadDirer, path string, walkFn filepath.WalkFunc, info fs.FileInfo, err error) error {
	if err != nil {
		return walkFn(path, info, err)
	}
	err = walkFn(path, info, nil)
	if !info.IsDir() {
		return err
	}
	if errors.Is(err, fs.SkipDir) {
		return nil
	}
	dirEntries, err := fileSystem.ReadDir(path)
	if err != nil {
		return err
	}
	sort.Sort(dirEntriesByName(dirEntries))
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		if name == "." || name == ".." {
			continue
		}
		info, err := dirEntry.Info()
		if err != nil {
			return err
		}
		if err := walk(fileSystem, filepath.Join(path, dirEntry.Name()), walkFn, info, nil); err != nil {
			return err
		}
	}
	return nil
}

// Walk is the equivalent of filepath.Walk but operates on fileSystem. Entries
// are returned in lexicographical order.
func Walk(fileSystem LstatReadDirer, path string, walkFn filepath.WalkFunc) error {
	info, err := fileSystem.Lstat(path)
	return walk(fileSystem, path, walkFn, info, err)
}

// WalkSlash is the equivalent of Walk but all paths are converted to use
// forward slashes with filepath.ToSlash.
func WalkSlash(fileSystem LstatReadDirer, path string, walkFn filepath.WalkFunc) error {
	return Walk(fileSystem, path, func(path string, info fs.FileInfo, err error) error {
		return walkFn(filepath.ToSlash(path), info, err)
	})
}
