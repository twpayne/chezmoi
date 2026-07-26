package chezmoi

import (
	"cmp"
	"path"
	"path/filepath"
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

// NewRelPath returns a new RelPath.
func NewRelPath(relPath string) RelPath {
	return RelPath{
		relPath: relPath,
	}
}

// NewRelPathFromComponents returns a new RelPath from the given components.
func NewRelPathFromComponents(components ...RelPath) RelPath {
	componentStrs := make([]string, len(components))
	for i, component := range components {
		componentStrs[i] = component.String()
	}
	return NewRelPath(strings.Join(componentStrs, "/"))
}

// NewUntrustedRelPath returns a clean version of path that does not escape its
// parent directory.
func NewUntrustedRelPath(relPath string) (RelPath, error) {
	if relPath == ".." || strings.HasPrefix(relPath, "../") || strings.Contains(relPath, "/../") || filepath.IsAbs(relPath) {
		return RelPath{}, InvalidPathError(relPath)
	}
	return NewRelPath(relPath), nil
}

// AppendSlash returns p with an appended slash.
func (p RelPath) AppendSlash() RelPath {
	return p.AppendString("/")
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

// Ext returns p's extension.
func (p RelPath) Ext() string {
	return path.Ext(p.relPath)
}

// HasDirPrefix returns true if p has dir prefix dirPrefix.
func (p RelPath) HasDirPrefix(dirPrefix RelPath) bool {
	return strings.HasPrefix(p.relPath, dirPrefix.String()+"/")
}

// IsEmpty returns if p is empty.
func (p RelPath) IsEmpty() bool {
	return p.relPath == ""
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
	strs := make([]string, len(ss)+1)
	strs[0] = p.relPath
	copy(strs[1:len(ss)+1], ss)
	return NewRelPath(path.Join(strs...))
}

// Len returns the length of p.
func (p RelPath) Len() int {
	return len(p.relPath)
}

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (p RelPath) MarshalText() ([]byte, error) {
	return []byte(p.relPath), nil
}

// Slice returns a part of p.
func (p RelPath) Slice(begin, end int) RelPath {
	return NewRelPath(p.relPath[begin:end])
}

// SourceRelDirPath returns p as a directory SourceRelPath.
func (p RelPath) SourceRelDirPath() SourceRelPath {
	return NewSourceRelDirPath(p.relPath)
}

// SourceRelPath returns p as a SourceRelPath.
func (p RelPath) SourceRelPath() SourceRelPath {
	return NewSourceRelPath(p.relPath)
}

// Split returns p's directory and path.
func (p RelPath) Split() (dirRelPath, fileRelPath RelPath) {
	dir, file := path.Split(p.relPath)
	return NewRelPath(dir), NewRelPath(file)
}

// SplitAll returns p's components.
func (p RelPath) SplitAll() []RelPath {
	components := strings.Split(p.relPath, "/")
	relPaths := make([]RelPath, len(components))
	for i, component := range components {
		relPaths[i] = NewRelPath(component)
	}
	return relPaths
}

func (p RelPath) String() string {
	return p.relPath
}

// TrimDirPrefix trims prefix from p.
func (p RelPath) TrimDirPrefix(dirPrefix RelPath) (RelPath, error) {
	if !p.HasDirPrefix(dirPrefix) {
		return EmptyRelPath, &NotInRelDirError{
			pathRelPath: p,
			dirRelPath:  dirPrefix,
		}
	}
	return p.Slice(dirPrefix.Len()+1, p.Len()), nil
}

// TrimTrailingSlash returns p with any trailing slash removed.
func (p RelPath) TrimTrailingSlash() RelPath {
	relPath, _ := strings.CutSuffix(p.relPath, "/")
	return NewRelPath(relPath)
}

// UnmarshalText implements encoding.TextUnmarshaler.UnmarshalText.
func (p *RelPath) UnmarshalText(data []byte) error {
	relPath, err := NewUntrustedRelPath(string(data))
	if err != nil {
		return err
	}
	*p = relPath
	return nil
}

// CompareRelPaths compares a and b.
func CompareRelPaths(a, b RelPath) int {
	return cmp.Compare(a.relPath, b.relPath)
}
