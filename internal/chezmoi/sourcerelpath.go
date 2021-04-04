package chezmoi

import (
	"strings"
)

// A SourceRelPath is a relative path to an entry in the source state.
type SourceRelPath struct {
	relPath RelPath
	isDir   bool
}

// SourceRelPaths is a slice of SourceRelPaths that implements sort.Interface.
type SourceRelPaths []SourceRelPath

func (ps SourceRelPaths) Len() int           { return len(ps) }
func (ps SourceRelPaths) Less(i, j int) bool { return string(ps[i].relPath) < string(ps[j].relPath) }
func (ps SourceRelPaths) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

// NewSourceRelDirPath returns a new SourceRelPath for a directory.
func NewSourceRelDirPath(relPath RelPath) SourceRelPath {
	return SourceRelPath{
		relPath: relPath,
		isDir:   true,
	}
}

// NewSourceRelPath returns a new SourceRelPath.
func NewSourceRelPath(relPath RelPath) SourceRelPath {
	return SourceRelPath{
		relPath: relPath,
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

// Join appends elems to p.
func (p SourceRelPath) Join(elems ...SourceRelPath) SourceRelPath {
	elemRelPaths := make(RelPaths, 0, len(elems))
	for _, elem := range elems {
		elemRelPaths = append(elemRelPaths, elem.relPath)
	}
	return SourceRelPath{
		relPath: p.relPath.Join(elemRelPaths...),
		isDir:   elems[len(elems)-1].isDir,
	}
}

// RelPath returns p as a relative path.
func (p SourceRelPath) RelPath() RelPath {
	return p.relPath
}

// Split returns the p's file and directory.
func (p SourceRelPath) Split() (SourceRelPath, SourceRelPath) {
	dir, file := p.relPath.Split()
	return NewSourceRelDirPath(dir), NewSourceRelPath(file)
}

func (p SourceRelPath) String() string {
	return string(p.relPath)
}

// TargetRelPath returns the relative path of p's target.
func (p SourceRelPath) TargetRelPath(encryptedSuffix string) RelPath {
	sourceNames := strings.Split(string(p.relPath), "/")
	relPathNames := make([]string, 0, len(sourceNames))
	if p.isDir {
		for _, sourceName := range sourceNames {
			dirAttr := parseDirAttr(sourceName)
			relPathNames = append(relPathNames, dirAttr.TargetName)
		}
	} else {
		for _, sourceName := range sourceNames[:len(sourceNames)-1] {
			dirAttr := parseDirAttr(sourceName)
			relPathNames = append(relPathNames, dirAttr.TargetName)
		}
		fileAttr := parseFileAttr(sourceNames[len(sourceNames)-1], encryptedSuffix)
		relPathNames = append(relPathNames, fileAttr.TargetName)
	}
	return RelPath(strings.Join(relPathNames, "/"))
}
