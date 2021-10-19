package chezmoi

import (
	"path"
	"strings"
)

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

// Ext returns p's extension.
func (p RelPath) Ext() string {
	return path.Ext(string(p))
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

func (p RelPath) String() string {
	return string(p)
}

// TrimDirPrefix trims prefix from p.
func (p RelPath) TrimDirPrefix(dirPrefix RelPath) (RelPath, error) {
	if !p.HasDirPrefix(dirPrefix) {
		return "", &notInRelDirError{
			pathRelPath: p,
			dirRelPath:  dirPrefix,
		}
	}
	return p[len(dirPrefix)+1:], nil
}
