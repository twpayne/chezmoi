package chezmoi

import "sort"

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
func (n *sourceStateEntryTreeNode) ForEach(targetRelPath RelPath, f func(RelPath, SourceStateEntry) error) error {
	return n.ForEachNode(targetRelPath, func(targetRelPath RelPath, node *sourceStateEntryTreeNode) error {
		if node.sourceStateEntry == nil {
			return nil
		}
		return f(targetRelPath, node.sourceStateEntry)
	})
}

// ForEachNode calls f for each node in the tree.
func (n *sourceStateEntryTreeNode) ForEachNode(targetRelPath RelPath, f func(RelPath, *sourceStateEntryTreeNode) error) error {
	if err := f(targetRelPath, n); err != nil {
		return err
	}

	childrenByRelPath := make(RelPaths, 0, len(n.children))
	for childRelPath := range n.children {
		childrenByRelPath = append(childrenByRelPath, childRelPath)
	}

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
