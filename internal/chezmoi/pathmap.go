package chezmoi

import (
	"fmt"

	"github.com/twpayne/chezmoi/v2/internal/chezmoimaps"
)

type (
	pathMapNode       any
	pathMapBranchNode map[RelPath]pathMapNode
	pathMapLeafNode   RelPath
)

// A PathMap maps selected paths from one directory hierarchy to another.
type PathMap struct {
	root pathMapNode
}

type pathMapError struct {
	fromRelPath RelPath
	toRelPath   RelPath
	reason      string
}

func (e *pathMapError) Error() string {
	return fmt.Sprintf("%s -> %s: %s", e.fromRelPath, e.toRelPath, e.reason)
}

// NewPathMap returns a new PathMap.
func NewPathMap() *PathMap {
	return &PathMap{
		root: make(pathMapBranchNode),
	}
}

// Add adds a mapping from fromRelPath to toRelPath. The mapping must be
// consistent with the existing mappings.
func (m *PathMap) Add(fromRelPath, toRelPath RelPath) error {
	node := m.root
	fromComponents := fromRelPath.SplitAll()
	for i, relPath := range fromComponents[:len(fromComponents)-1] {
		switch nodeConcrete := node.(type) {
		case pathMapBranchNode:
			child, ok := nodeConcrete[relPath]
			if !ok {
				child = make(pathMapBranchNode)
				nodeConcrete[relPath] = child
			}
			node = child
		case pathMapLeafNode:
			return &pathMapError{
				fromRelPath: fromRelPath,
				toRelPath:   toRelPath,
				reason:      fmt.Sprintf("%s is already mapped", EmptyRelPath.Join(fromComponents[:i]...)),
			}
		}
	}

	lastComponent := fromComponents[len(fromComponents)-1]
	switch node := node.(type) {
	case pathMapBranchNode:
		if _, ok := node[lastComponent]; ok {
			return &pathMapError{
				fromRelPath: fromRelPath,
				toRelPath:   toRelPath,
				reason:      "directory is already mapped",
			}
		}
		node[lastComponent] = pathMapLeafNode(toRelPath)
	case pathMapLeafNode:
		if RelPath(node) != toRelPath {
			return &pathMapError{
				fromRelPath: fromRelPath,
				toRelPath:   toRelPath,
				reason:      fmt.Sprintf("parent %s is already mapped to %s", EmptyRelPath.Join(fromComponents[:len(fromComponents)-1]...), RelPath(node)),
			}
		}
	}

	return nil
}

func (m *PathMap) AddStringMap(stringMap map[string]string) error {
	for _, key := range chezmoimaps.SortedKeys(stringMap) {
		if err := m.Add(NewRelPath(key), NewRelPath(stringMap[key])); err != nil {
			return err
		}
	}
	return nil
}

// Lookup looks up the mapping for fromRelPath. If a prefix of fromRelPath is in
// m then it returns the corresponding mapping. Otherwise, it returns
// fromRelPath unchanged.
func (m *PathMap) Lookup(fromRelPath RelPath) RelPath {
	node := m.root
	fromComponents := fromRelPath.SplitAll()
	for i, relPath := range fromComponents {
		switch nodeConcrete := node.(type) {
		case pathMapBranchNode:
			child, ok := nodeConcrete[relPath]
			if !ok {
				return fromRelPath
			}
			node = child
		case pathMapLeafNode:
			return RelPath(nodeConcrete).Join(fromComponents[i:]...)
		}
	}
	switch node := node.(type) {
	case pathMapBranchNode:
		return fromRelPath
	case pathMapLeafNode:
		return RelPath(node)
	default:
		panic("unreachable")
	}
}
