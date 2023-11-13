package chezmoi

import (
	"path"
	"strings"
)

var (
	DotRelPath   = NewRelPath(".")
	EmptyRelPath = NewRelPath("")
)

// A RelPath is a relative path.
type RelPath struct {
	relPath string
}

// RelPaths is a slice of RelPaths that implements sort.Interface.
type RelPaths []RelPath

func (ps RelPaths) Len() int           { return len(ps) }
func (ps RelPaths) Less(i, j int) bool { return ps[i].Less(ps[j]) }
func (ps RelPaths) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

// NewRelPath returns a new RelPath.
func NewRelPath(relPath string) RelPath {
	return RelPath{
		relPath: relPath,
	}
}

// AppendString appends s to p.
func (p RelPath) AppendString(s string) RelPath {
	return NewRelPath(p.relPath + s)
}

// Base returns p's base name.
func (p RelPath) Base() string {
	return path.Base(p.relPath)
}

// Dir returns p's directory.
func (p RelPath) Dir() RelPath {
	return NewRelPath(path.Dir(p.relPath))
}

// Empty returns if p is empty.
func (p RelPath) Empty() bool {
	return p.relPath == ""
}

// Ext returns p's extension.
func (p RelPath) Ext() string {
	return path.Ext(p.relPath)
}

// HasDirPrefix returns true if p has dir prefix dirPrefix.
func (p RelPath) HasDirPrefix(dirPrefix RelPath) bool {
	return strings.HasPrefix(p.relPath, dirPrefix.String()+"/")
}

// Join appends relPaths to p.
func (p RelPath) Join(relPaths ...RelPath) RelPath {
	relPathStrs := make([]string, 0, len(relPaths)+1)
	if p.relPath != "" {
		relPathStrs = append(relPathStrs, p.relPath)
	}
	for _, relPath := range relPaths {
		relPathStrs = append(relPathStrs, relPath.String())
	}
	return NewRelPath(path.Join(relPathStrs...))
}

// JoinString returns a new RelPath with ss appended.
func (p RelPath) JoinString(ss ...string) RelPath {
	strs := make([]string, 0, len(ss)+1)
	strs = append(strs, p.relPath)
	strs = append(strs, ss...)
	return NewRelPath(path.Join(strs...))
}

// Len returns the length of p.
func (p RelPath) Len() int {
	return len(p.relPath)
}

// Less returns true if p is less than other.
func (p RelPath) Less(other RelPath) bool {
	return p.relPath < other.relPath
}

// MarshalJSON implements encoding.TextMarshaler.MarshalJSON.
func (p RelPath) MarshalJSON() ([]byte, error) {
	return FormatJSON.Marshal(p.relPath)
}

// Slice returns a part of p.
func (p RelPath) Slice(begin, end int) RelPath {
	return NewRelPath(p.relPath[begin:end])
}

// SourceRelPath returns p as a SourceRelPath.
func (p RelPath) SourceRelPath() SourceRelPath {
	return NewSourceRelPath(p.relPath)
}

// SourceRelDirPath returns p as a directory SourceRelPath.
func (p RelPath) SourceRelDirPath() SourceRelPath {
	return NewSourceRelDirPath(p.relPath)
}

// Split returns p's directory and path.
func (p RelPath) Split() (RelPath, RelPath) {
	dir, file := path.Split(p.relPath)
	return NewRelPath(dir), NewRelPath(file)
}

// SplitAll returns p's components.
func (p RelPath) SplitAll() []RelPath {
	components := strings.Split(p.relPath, "/")
	relPaths := make([]RelPath, 0, len(components))
	for _, component := range components {
		relPaths = append(relPaths, NewRelPath(component))
	}
	return relPaths
}

func (p RelPath) String() string {
	return p.relPath
}

// TrimDirPrefix trims prefix from p.
func (p RelPath) TrimDirPrefix(dirPrefix RelPath) (RelPath, error) {
	if !p.HasDirPrefix(dirPrefix) {
		return EmptyRelPath, &notInRelDirError{
			pathRelPath: p,
			dirRelPath:  dirPrefix,
		}
	}
	return p.Slice(dirPrefix.Len()+1, p.Len()), nil
}
