package chezmoi

import (
	"context"
	"sort"

	"golang.org/x/sync/errgroup"
)

// A concurrentTreeNode is a tree that can be walked concurrently, with parents
// always visited before their children.
type concurrentTreeNode map[RelPath]concurrentTreeNode

type ConcurrentTree struct {
	relPathSet relPathSet
	root       concurrentTreeNode
}

// NewConcurrentTree returns a new ConcurrentTree from relPaths.
func NewConcurrentTree(relPaths RelPaths) *ConcurrentTree {
	sort.Sort(relPaths)
	root := make(concurrentTreeNode)
	for _, relPath := range relPaths {
		root.add(relPath.SplitAll())
	}
	return &ConcurrentTree{
		relPathSet: newRelPathSet(relPaths),
		root:       root,
	}
}

func (t *ConcurrentTree) WalkChildren(ctx context.Context, relPath RelPath, f func(RelPath) error) error {
	return t.root.walkChildren(ctx, relPath, func(relPath RelPath) error {
		if !t.relPathSet.contains(relPath) {
			return nil
		}
		return f(relPath)
	})
}

// add adds the RelPath composed of relPathComponents to n.
func (n concurrentTreeNode) add(relPathComponents []RelPath) {
	child, ok := n[relPathComponents[0]]
	if !ok {
		child = make(concurrentTreeNode)
		n[relPathComponents[0]] = child
	}
	if len(relPathComponents) > 1 {
		child.add(relPathComponents[1:])
	}
}

// Walk walks n concurrently.
func (n concurrentTreeNode) walk(ctx context.Context, relPath RelPath, f func(RelPath) error) error {
	if err := f(relPath); err != nil {
		return err
	}
	return n.walkChildren(ctx, relPath, f)
}

// WalkChildren walks n's children concurrently.
func (n concurrentTreeNode) walkChildren(ctx context.Context, relPath RelPath, f func(RelPath) error) error {
	if len(n) == 0 {
		return nil
	}
	group, ctx := errgroup.WithContext(ctx)
	// FIXME IAMHERE
	// the idea is that we visit entries within the directory in a deterministic order
	// so we sort on component (or targetRelPath, which I think is already the case) and then call f
	// the complexity comes from handling run_before_/run_after_ scripts, and by the time we're here we've lost the before/after information
	// need to maintain it somehow
	for childRelPathComponent, child := range n {
		childRelPath := relPath.Join(childRelPathComponent)
		child := child
		walkChildFunc := func() error {
			return child.walk(ctx, childRelPath, f)
		}
		group.Go(walkChildFunc)
	}
	return group.Wait()
}
