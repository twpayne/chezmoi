package chezmoi

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"golang.org/x/exp/maps"
)

// A sourceStateEntryTreeNode is a node in a tree of SourceStateEntries.
type sourceStateEntryTreeNode struct {
	sourceStateEntry SourceStateEntry
	children         map[RelPath]*sourceStateEntryTreeNode
}

// newSourceStateTreeNode returns a new sourceStateEntryTreeNode.
func newSourceStateTreeNode() *sourceStateEntryTreeNode {
	return &sourceStateEntryTreeNode{}
}

// Get returns the SourceStateEntry at relPath.
func (n *sourceStateEntryTreeNode) Get(relPath RelPath) SourceStateEntry {
	node := n.GetNode(relPath)
	if node == nil {
		return nil
	}
	return node.sourceStateEntry
}

// GetNode returns the SourceStateTreeNode at relPath.
func (n *sourceStateEntryTreeNode) GetNode(targetRelPath RelPath) *sourceStateEntryTreeNode {
	if targetRelPath.Empty() {
		return n
	}

	node := n
	for _, childRelPath := range targetRelPath.SplitAll() {
		var ok bool
		node, ok = node.children[childRelPath]
		if !ok {
			return nil
		}
	}
	return node
}

// ForEach calls f for each SourceStateEntry in the tree.
func (n *sourceStateEntryTreeNode) ForEach(
	targetRelPath RelPath,
	f func(RelPath, SourceStateEntry) error,
) error {
	return n.ForEachNode(
		targetRelPath,
		func(targetRelPath RelPath, node *sourceStateEntryTreeNode) error {
			if node.sourceStateEntry == nil {
				return nil
			}
			return f(targetRelPath, node.sourceStateEntry)
		},
	)
}

// ForEachNode calls f for each node in the tree.
func (n *sourceStateEntryTreeNode) ForEachNode(
	targetRelPath RelPath, f func(RelPath, *sourceStateEntryTreeNode) error,
) error {
	switch err := f(targetRelPath, n); {
	case errors.Is(err, fs.SkipDir):
		return nil
	case err != nil:
		return err
	}

	childrenByRelPath := RelPaths(maps.Keys(n.children))
	sort.Sort(childrenByRelPath)
	for _, childRelPath := range childrenByRelPath {
		child := n.children[childRelPath]
		if err := child.ForEachNode(targetRelPath.Join(childRelPath), f); err != nil {
			return err
		}
	}

	return nil
}

// Map returns a map of relPaths to SourceStateEntries.
func (n *sourceStateEntryTreeNode) Map() map[RelPath]SourceStateEntry {
	m := make(map[RelPath]SourceStateEntry)
	_ = n.ForEach(EmptyRelPath, func(relPath RelPath, sourceStateEntry SourceStateEntry) error {
		m[relPath] = sourceStateEntry
		return nil
	})
	return m
}

// MkdirAll creates SourceStateDirs for all components of targetRelPath if they
// do not already exist and returns the SourceStateDir of relPath.
func (n *sourceStateEntryTreeNode) MkdirAll(
	targetRelPath RelPath, origin SourceStateOrigin, umask fs.FileMode,
) (*SourceStateDir, error) {
	if targetRelPath == EmptyRelPath {
		return nil, nil
	}

	node := n
	var sourceRelPath SourceRelPath
	componentRelPaths := targetRelPath.SplitAll()
	var sourceStateDir *SourceStateDir
	for i, componentRelPath := range componentRelPaths {
		if node.children == nil {
			node.children = make(map[RelPath]*sourceStateEntryTreeNode)
		}
		if child, ok := node.children[componentRelPath]; ok {
			node = child
		} else {
			child = newSourceStateTreeNode()
			node.children[componentRelPath] = child
			node = child
		}

		switch {
		case node.sourceStateEntry == nil:
			dirAttr := DirAttr{
				TargetName: componentRelPath.String(),
			}
			targetStateDir := &TargetStateDir{
				perm: dirAttr.perm() &^ umask,
			}
			sourceRelPath = sourceRelPath.Join(NewSourceRelPath(dirAttr.SourceName()))
			sourceStateDir = &SourceStateDir{
				Attr:             dirAttr,
				origin:           origin,
				sourceRelPath:    sourceRelPath,
				targetStateEntry: targetStateDir,
			}
			node.sourceStateEntry = sourceStateDir
		default:
			var ok bool
			sourceStateDir, ok = node.sourceStateEntry.(*SourceStateDir)
			if !ok {
				return nil, fmt.Errorf(
					"%s: not a directory",
					componentRelPaths[0].Join(componentRelPaths[1:i+1]...),
				)
			}
			sourceRelPath = sourceRelPath.Join(NewSourceRelPath(sourceStateDir.Attr.SourceName()))
		}
	}
	return sourceStateDir, nil
}

// Set sets the SourceStateEntry at relPath to sourceStateEntry.
func (n *sourceStateEntryTreeNode) Set(targetRelPath RelPath, sourceStateEntry SourceStateEntry) {
	if targetRelPath.Empty() {
		n.sourceStateEntry = sourceStateEntry
		return
	}

	node := n
	for _, childRelPath := range targetRelPath.SplitAll() {
		if node.children == nil {
			node.children = make(map[RelPath]*sourceStateEntryTreeNode)
		}
		if child, ok := node.children[childRelPath]; ok {
			node = child
		} else {
			child = newSourceStateTreeNode()
			node.children[childRelPath] = child
			node = child
		}
	}
	node.sourceStateEntry = sourceStateEntry
}
