package chezmoi

import (
	"path"
	"strings"
)

var emptySourceRelPath SourceRelPath

// A SourceRelPath is a relative path to an entry in the source state.
type SourceRelPath struct {
	relPath RelPath
	isDir   bool
}

// NewSourceRelDirPath returns a new SourceRelPath for a directory.
func NewSourceRelDirPath(relPath string) SourceRelPath {
	return SourceRelPath{
		relPath: NewRelPath(relPath),
		isDir:   true,
	}
}

// NewSourceRelPath returns a new SourceRelPath.
func NewSourceRelPath(relPath string) SourceRelPath {
	return SourceRelPath{
		relPath: NewRelPath(relPath),
	}
}

// Dir returns p's directory.
func (p SourceRelPath) Dir() SourceRelPath {
	return SourceRelPath{
		relPath: p.relPath.Dir(),
		isDir:   true,
	}
}

// Empty returns true if p is empty.
func (p SourceRelPath) Empty() bool {
	return p == SourceRelPath{}
}

// Join appends sourceRelPaths to p.
func (p SourceRelPath) Join(sourceRelPaths ...SourceRelPath) SourceRelPath {
	relPaths := make([]RelPath, 0, len(sourceRelPaths))
	for _, sourceRelPath := range sourceRelPaths {
		relPaths = append(relPaths, sourceRelPath.relPath)
	}
	return SourceRelPath{
		relPath: p.relPath.Join(relPaths...),
		isDir:   sourceRelPaths[len(sourceRelPaths)-1].isDir,
	}
}

// Less returns true if p is less than other.
func (p SourceRelPath) Less(other SourceRelPath) bool {
	return p.relPath.Less(other.relPath)
}

// RelPath returns p as a relative path.
func (p SourceRelPath) RelPath() RelPath {
	return p.relPath
}

// Split returns the p's file and directory.
func (p SourceRelPath) Split() (SourceRelPath, SourceRelPath) {
	dir, file := p.relPath.Split()
	return NewSourceRelDirPath(dir.String()), NewSourceRelPath(file.String())
}

func (p SourceRelPath) String() string {
	return p.relPath.String()
}

// TargetRelPath returns the relative path of p's target.
func (p SourceRelPath) TargetRelPath(encryption Encryption) RelPath {
	sourceNames := strings.Split(p.relPath.String(), "/")
	relPathStrs := make([]string, 0, len(sourceNames))
	if p.isDir {
		for _, sourceName := range sourceNames {
			dirAttr := parseDirAttr(sourceName)
			relPathStrs = append(relPathStrs, dirAttr.TargetName)
		}
	} else {
		for _, sourceName := range sourceNames[:len(sourceNames)-1] {
			dirAttr := parseDirAttr(sourceName)
			relPathStrs = append(relPathStrs, dirAttr.TargetName)
		}
		fileAttr := parseFileAttr(sourceNames[len(sourceNames)-1], encryption)
		relPathStrs = append(relPathStrs, fileAttr.TargetName)
	}
	return NewRelPath(path.Join(relPathStrs...))
}
