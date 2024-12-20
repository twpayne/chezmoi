package cmd

import (
	"maps"
	"slices"
	"strings"
)

type pathListTreeNode struct {
	component string
	children  map[string]*pathListTreeNode
}

func newPathListTreeNode(component string) *pathListTreeNode {
	return &pathListTreeNode{
		component: component,
		children:  make(map[string]*pathListTreeNode),
	}
}

func newPathListTreeFromPathsSlice(paths []string) *pathListTreeNode {
	root := newPathListTreeNode("")
	for _, path := range paths {
		n := root
		for _, component := range strings.Split(path, "/") {
			child, ok := n.children[component]
			if !ok {
				child = newPathListTreeNode(component)
				n.children[component] = child
			}
			n = child
		}
	}
	return root
}

func (n *pathListTreeNode) write(sb *strings.Builder, prefix, indent string) {
	sb.WriteString(prefix)
	sb.WriteString(n.component)
	sb.WriteByte('\n')
	n.writeChildren(sb, prefix+indent, indent)
}

func (n *pathListTreeNode) writeChildren(sb *strings.Builder, prefix, indent string) {
	for _, key := range slices.Sorted(maps.Keys(n.children)) {
		child := n.children[key]
		child.write(sb, prefix, indent)
	}
}
