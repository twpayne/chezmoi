package chezmoi

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
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

// get returns the SourceStateEntry at relPath.
func (n *sourceStateEntryTreeNode) get(relPath RelPath) SourceStateEntry {
	nodes := n.getNodes(relPath)
	if nodes == nil {
		return nil
	}
	return nodes[len(nodes)-1].sourceStateEntry
}

// getNodes returns the sourceStateEntryTreeNodes to reach targetRelPath.
func (n *sourceStateEntryTreeNode) getNodes(targetRelPath RelPath) []*sourceStateEntryTreeNode {
	if targetRelPath.Empty() {
		return []*sourceStateEntryTreeNode{n}
	}

	targetRelPathComponents := targetRelPath.SplitAll()
	nodes := make([]*sourceStateEntryTreeNode, 0, len(targetRelPathComponents))
	nodes = append(nodes, n)
	for _, childRelPath := range targetRelPathComponents {
		childNode, ok := nodes[len(nodes)-1].children[childRelPath]
		if !ok {
			return nil
		}
		nodes = append(nodes, childNode)
	}
	return nodes
}

// forEach calls f for each SourceStateEntry in the tree.
func (n *sourceStateEntryTreeNode) forEach(targetRelPath RelPath, f func(RelPath, SourceStateEntry) error) error {
	return n.forEachNode(targetRelPath, func(targetRelPath RelPath, node *sourceStateEntryTreeNode) error {
		if node.sourceStateEntry == nil {
			return nil
		}
		return f(targetRelPath, node.sourceStateEntry)
	})
}

// forEachNode calls f for each node in the tree.
func (n *sourceStateEntryTreeNode) forEachNode(targetRelPath RelPath, f func(RelPath, *sourceStateEntryTreeNode) error) error {
	switch err := f(targetRelPath, n); {
	case errors.Is(err, fs.SkipDir):
		return nil
	case err != nil:
		return err
	}

	childrenByRelPath := slices.Collect(maps.Keys(n.children))
	slices.SortFunc(childrenByRelPath, CompareRelPaths)
	for _, childRelPath := range childrenByRelPath {
		child := n.children[childRelPath]
		if err := child.forEachNode(targetRelPath.Join(childRelPath), f); err != nil {
			return err
		}
	}

	return nil
}

// getMap returns a map of relPaths to SourceStateEntries.
func (n *sourceStateEntryTreeNode) getMap() map[RelPath]SourceStateEntry {
	m := make(map[RelPath]SourceStateEntry)
	_ = n.forEach(EmptyRelPath, func(relPath RelPath, sourceStateEntry SourceStateEntry) error {
		m[relPath] = sourceStateEntry
		return nil
	})
	return m
}

// mkdirAll creates SourceStateDirs for all components of targetRelPath if they
// do not already exist and returns the SourceStateDir of relPath.
func (n *sourceStateEntryTreeNode) mkdirAll(
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

		if node.sourceStateEntry == nil {
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
		} else {
			var ok bool
			sourceStateDir, ok = node.sourceStateEntry.(*SourceStateDir)
			if !ok {
				return nil, fmt.Errorf("%s: not a directory", componentRelPaths[0].Join(componentRelPaths[1:i+1]...))
			}
			sourceRelPath = sourceRelPath.Join(NewSourceRelPath(sourceStateDir.Attr.SourceName()))
		}
	}
	return sourceStateDir, nil
}

// set sets the SourceStateEntry at relPath to sourceStateEntry.
func (n *sourceStateEntryTreeNode) set(targetRelPath RelPath, sourceStateEntry SourceStateEntry) {
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
