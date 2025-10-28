package chezmoi

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
)

// A SourceStateEntryTreeNode is a node in a tree of SourceStateEntries.
type SourceStateEntryTreeNode struct {
	SourceStateEntry SourceStateEntry
	Children         map[RelPath]*SourceStateEntryTreeNode
}

// NewSourceStateEntryTreeNode returns a new sourceStateEntryTreeNode.
func NewSourceStateEntryTreeNode() *SourceStateEntryTreeNode {
	return &SourceStateEntryTreeNode{}
}

// Get returns the SourceStateEntry at relPath.
func (n *SourceStateEntryTreeNode) Get(relPath RelPath) SourceStateEntry {
	nodes := n.GetNodes(relPath)
	if nodes == nil {
		return nil
	}
	return nodes[len(nodes)-1].SourceStateEntry
}

// GetNodes returns the sourceStateEntryTreeNodes to reach targetRelPath.
func (n *SourceStateEntryTreeNode) GetNodes(targetRelPath RelPath) []*SourceStateEntryTreeNode {
	if targetRelPath.Empty() {
		return []*SourceStateEntryTreeNode{n}
	}

	targetRelPathComponents := targetRelPath.SplitAll()
	nodes := make([]*SourceStateEntryTreeNode, 0, len(targetRelPathComponents))
	nodes = append(nodes, n)
	for _, childRelPath := range targetRelPathComponents {
		childNode, ok := nodes[len(nodes)-1].Children[childRelPath]
		if !ok {
			return nil
		}
		nodes = append(nodes, childNode)
	}
	return nodes
}

// ForEach calls f for each SourceStateEntry in the tree.
func (n *SourceStateEntryTreeNode) ForEach(targetRelPath RelPath, f func(RelPath, SourceStateEntry) error) error {
	return n.ForEachNode(targetRelPath, func(targetRelPath RelPath, node *SourceStateEntryTreeNode) error {
		if node.SourceStateEntry == nil {
			return nil
		}
		return f(targetRelPath, node.SourceStateEntry)
	})
}

// ForEachNode calls f for each node in the tree.
func (n *SourceStateEntryTreeNode) ForEachNode(targetRelPath RelPath, f func(RelPath, *SourceStateEntryTreeNode) error) error {
	switch err := f(targetRelPath, n); {
	case errors.Is(err, fs.SkipDir):
		return nil
	case err != nil:
		return err
	}

	childrenByRelPath := slices.Collect(maps.Keys(n.Children))
	slices.SortFunc(childrenByRelPath, CompareRelPaths)
	for _, childRelPath := range childrenByRelPath {
		child := n.Children[childRelPath]
		if err := child.ForEachNode(targetRelPath.Join(childRelPath), f); err != nil {
			return err
		}
	}

	return nil
}

// GetMap returns a map of relPaths to SourceStateEntries.
func (n *SourceStateEntryTreeNode) GetMap() map[RelPath]SourceStateEntry {
	m := make(map[RelPath]SourceStateEntry)
	_ = n.ForEach(EmptyRelPath, func(relPath RelPath, sourceStateEntry SourceStateEntry) error {
		m[relPath] = sourceStateEntry
		return nil
	})
	return m
}

// MkdirAll creates SourceStateDirs for all components of targetRelPath if they
// do not already exist and returns the SourceStateDir of relPath.
func (n *SourceStateEntryTreeNode) MkdirAll(
	targetRelPath RelPath,
	origin SourceStateOrigin,
	umask fs.FileMode,
) (*SourceStateDir, error) {
	if targetRelPath == EmptyRelPath {
		return nil, nil
	}

	node := n
	var sourceRelPath SourceRelPath
	componentRelPaths := targetRelPath.SplitAll()
	var sourceStateDir *SourceStateDir
	for i, componentRelPath := range componentRelPaths {
		if node.Children == nil {
			node.Children = make(map[RelPath]*SourceStateEntryTreeNode)
		}
		if child, ok := node.Children[componentRelPath]; ok {
			node = child
		} else {
			child = NewSourceStateEntryTreeNode()
			node.Children[componentRelPath] = child
			node = child
		}

		if node.SourceStateEntry == nil {
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
			node.SourceStateEntry = sourceStateDir
		} else {
			var ok bool
			sourceStateDir, ok = node.SourceStateEntry.(*SourceStateDir)
			if !ok {
				return nil, fmt.Errorf("%s: not a directory", componentRelPaths[0].Join(componentRelPaths[1:i+1]...))
			}
			sourceRelPath = sourceRelPath.Join(NewSourceRelPath(sourceStateDir.Attr.SourceName()))
		}
	}
	return sourceStateDir, nil
}

// Set sets the SourceStateEntry at relPath to sourceStateEntry.
func (n *SourceStateEntryTreeNode) Set(targetRelPath RelPath, sourceStateEntry SourceStateEntry) {
	if targetRelPath.Empty() {
		n.SourceStateEntry = sourceStateEntry
		return
	}

	node := n
	for _, childRelPath := range targetRelPath.SplitAll() {
		if node.Children == nil {
			node.Children = make(map[RelPath]*SourceStateEntryTreeNode)
		}
		if child, ok := node.Children[childRelPath]; ok {
			node = child
		} else {
			child = NewSourceStateEntryTreeNode()
			node.Children[childRelPath] = child
			node = child
		}
	}
	node.SourceStateEntry = sourceStateEntry
}
