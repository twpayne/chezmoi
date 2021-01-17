package chezmoi

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// An AbsPath is an absolute path.
type AbsPath string

// NewAbsPath returns a new AbsPath.
func NewAbsPath(path string) (AbsPath, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("%s: not an absolute path", path)
	}
	return AbsPath(path), nil
}

// Base returns p's basename.
func (p AbsPath) Base() string {
	return path.Base(string(p))
}

// Dir returns p's directory.
func (p AbsPath) Dir() AbsPath {
	return AbsPath(path.Dir(string(p)))
}

// Join appends elems to p.
func (p AbsPath) Join(elems ...RelPath) AbsPath {
	elemStrs := make([]string, 0, len(elems)+1)
	elemStrs = append(elemStrs, string(p))
	for _, elem := range elems {
		elemStrs = append(elemStrs, string(elem))
	}
	return AbsPath(path.Join(elemStrs...))
}

// MustTrimDirPrefix is like TrimPrefix but panics on any error.
func (p AbsPath) MustTrimDirPrefix(dirPrefix AbsPath) RelPath {
	relPath, err := p.TrimDirPrefix(dirPrefix)
	if err != nil {
		panic(err)
	}
	return relPath
}

// Split returns p's directory and file.
func (p AbsPath) Split() (AbsPath, RelPath) {
	dir, file := path.Split(string(p))
	return AbsPath(dir), RelPath(file)
}

// TrimDirPrefix trims prefix from p.
func (p AbsPath) TrimDirPrefix(dirPrefixAbsPath AbsPath) (RelPath, error) {
	if !strings.HasPrefix(string(p), string(dirPrefixAbsPath+"/")) {
		return "", &errNotInAbsDir{
			pathAbsPath: p,
			dirAbsPath:  dirPrefixAbsPath,
		}
	}
	return RelPath(p[len(dirPrefixAbsPath)+1:]), nil
}

// AbsPaths is a slice of AbsPaths that implements sort.Interface.
type AbsPaths []AbsPath

func (ps AbsPaths) Len() int           { return len(ps) }
func (ps AbsPaths) Less(i, j int) bool { return string(ps[i]) < string(ps[j]) }
func (ps AbsPaths) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

// A RelPath is a relative path.
type RelPath string

// Base returns p's base name.
func (p RelPath) Base() string {
	return path.Base(string(p))
}

// Dir returns p's directory.
func (p RelPath) Dir() RelPath {
	return RelPath(path.Dir(string(p)))
}

// HasDirPrefix returns true if p has dir prefix dirPrefix.
func (p RelPath) HasDirPrefix(dirPrefix RelPath) bool {
	return strings.HasPrefix(string(p), string(dirPrefix)+"/")
}

// Join appends elems to p.
func (p RelPath) Join(elems ...RelPath) RelPath {
	elemStrs := make([]string, 0, len(elems)+1)
	elemStrs = append(elemStrs, string(p))
	for _, elem := range elems {
		elemStrs = append(elemStrs, string(elem))
	}
	return RelPath(path.Join(elemStrs...))
}

// Split returns p's directory and path.
func (p RelPath) Split() (RelPath, RelPath) {
	dir, file := path.Split(string(p))
	return RelPath(dir), RelPath(file)
}

// TrimDirPrefix trims prefix from p.
func (p RelPath) TrimDirPrefix(dirPrefix RelPath) (RelPath, error) {
	if !p.HasDirPrefix(dirPrefix) {
		return "", &errNotInRelDir{
			pathRelPath: p,
			dirRelPath:  dirPrefix,
		}
	}
	return p[len(dirPrefix)+1:], nil
}

// RelPaths is a slice of RelPaths that implements sort.Interface.
type RelPaths []RelPath

func (ps RelPaths) Len() int           { return len(ps) }
func (ps RelPaths) Less(i, j int) bool { return string(ps[i]) < string(ps[j]) }
func (ps RelPaths) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
