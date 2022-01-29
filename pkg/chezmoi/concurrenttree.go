package chezmoi

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/sync/errgroup"
)

// A ConcurrentTree is a tree that can be walked concurrently, with parents
// always visited before their children.
type ConcurrentTree map[RelPath]ConcurrentTree

// NewConcurrentTree returns a new ConcurrentTree from relPaths.
func NewConcurrentTree(relPaths RelPaths) ConcurrentTree {
	sort.Sort(relPaths)
	root := make(ConcurrentTree)
	for _, relPath := range relPaths {
		// FIXME remove
		if strings.Contains(relPath.String(), ".chezmoiscripts") {
			panic(fmt.Sprintf("found .chezmoiscripts at %s", relPath))
		}
		root.add(relPath.SplitAll())
	}
	return root
}

// Walk walks n concurrently.
func (n ConcurrentTree) Walk(ctx context.Context, relPath RelPath, f func(RelPath) error) error {
	if err := f(relPath); err != nil {
		return err
	}
	return n.WalkChildren(ctx, relPath, f)
}

// WalkChildren walks n's children concurrently.
func (n ConcurrentTree) WalkChildren(ctx context.Context, relPath RelPath, f func(RelPath) error) error {
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
			return child.Walk(ctx, childRelPath, f)
		}
		group.Go(walkChildFunc)
	}
	return group.Wait()
}

// add adds the RelPath composed of relPathComponents to n.
func (n ConcurrentTree) add(relPathComponents []RelPath) {
	child, ok := n[relPathComponents[0]]
	if !ok {
		child = make(ConcurrentTree)
		n[relPathComponents[0]] = child
	}
	if len(relPathComponents) > 1 {
		child.add(relPathComponents[1:])
	}
}
