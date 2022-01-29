package chezmoi

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConcurrentTree(t *testing.T) {
	for i, tc := range []struct {
		relPaths RelPaths
		expected *ConcurrentTree
	}{
		{
			relPaths: nil,
			expected: &ConcurrentTree{
				root: concurrentTreeNode{},
			},
		},
		{
			relPaths: RelPaths{
				NewRelPath("dir"),
			},
			expected: &ConcurrentTree{
				root: concurrentTreeNode{
					NewRelPath("dir"): concurrentTreeNode{},
				},
			},
		},
		{
			relPaths: RelPaths{
				NewRelPath("dir"),
				NewRelPath("dir/file"),
			},
			expected: &ConcurrentTree{
				root: concurrentTreeNode{
					NewRelPath("dir"): concurrentTreeNode{
						NewRelPath("file"): concurrentTreeNode{},
					},
				},
			},
		},
		{
			relPaths: RelPaths{
				NewRelPath("dir"),
				NewRelPath("dir/file"),
				NewRelPath("dir/subdir"),
				NewRelPath("dir/subdir/file"),
			},
			expected: &ConcurrentTree{
				root: concurrentTreeNode{
					NewRelPath("dir"): concurrentTreeNode{
						NewRelPath("file"): concurrentTreeNode{},
						NewRelPath("subdir"): concurrentTreeNode{
							NewRelPath("file"): concurrentTreeNode{},
						},
					},
				},
			},
		},
		{
			relPaths: RelPaths{
				NewRelPath("dir"),
				NewRelPath("dir/file"),
				NewRelPath("dir/subdir"),
				NewRelPath("dir/subdir/file"),
				NewRelPath("file"),
			},
			expected: &ConcurrentTree{
				root: concurrentTreeNode{
					NewRelPath("dir"): concurrentTreeNode{
						NewRelPath("file"): concurrentTreeNode{},
						NewRelPath("subdir"): concurrentTreeNode{
							NewRelPath("file"): concurrentTreeNode{},
						},
					},
					NewRelPath("file"): concurrentTreeNode{},
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := NewConcurrentTree(tc.relPaths)
			tc.expected.relPathSet = newRelPathSet(tc.relPaths)
			require.Equal(t, tc.expected, actual)
			ctx := context.Background()
			var visitedRelPathMu sync.Mutex
			var visitedRelPaths RelPaths
			require.NoError(t, actual.WalkChildren(ctx, EmptyRelPath, func(relPath RelPath) error {
				visitedRelPathMu.Lock()
				visitedRelPaths = append(visitedRelPaths, relPath)
				visitedRelPathMu.Unlock()
				return nil
			}))
			sort.Sort(visitedRelPaths)
			assert.Equal(t, tc.relPaths, visitedRelPaths)
		})
	}
}

func TestConcurrentTreeWalkChildren(t *testing.T) {
	for i, tc := range []struct {
		relPaths RelPaths
		expected concurrentTreeNode
	}{
		{
			relPaths: RelPaths{
				NewRelPath("dir/file"),
				NewRelPath("dir/subdir"),
				NewRelPath("dir/subdir/file"),
				NewRelPath("file"),
			},
		},
		{
			relPaths: RelPaths{
				NewRelPath("dir/file"),
				NewRelPath("dir/subdir/file"),
				NewRelPath("file"),
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := NewConcurrentTree(tc.relPaths)
			ctx := context.Background()
			var visitedRelPathMu sync.Mutex
			var visitedRelPaths RelPaths
			require.NoError(t, actual.WalkChildren(ctx, EmptyRelPath, func(relPath RelPath) error {
				visitedRelPathMu.Lock()
				visitedRelPaths = append(visitedRelPaths, relPath)
				visitedRelPathMu.Unlock()
				return nil
			}))
			sort.Sort(visitedRelPaths)
			assert.Equal(t, tc.relPaths, visitedRelPaths)
		})
	}
}

func TestConcurrentTreeWalkError(t *testing.T) {
	ct := NewConcurrentTree(RelPaths{
		NewRelPath("dir"),
		NewRelPath("dir/file"),
		NewRelPath("dir/subdir/file"),
		NewRelPath("dir/subdir2"),
		NewRelPath("dir/subdir2/file"),
		NewRelPath("file"),
	})

	// Walk, but return an error for dir/subdir2/file.
	var visitedRelPathshMu sync.Mutex
	visitedRelPaths := make(map[RelPath]bool)
	walkErr := errors.New("walk")
	require.Equal(t, walkErr, ct.WalkChildren(context.Background(), EmptyRelPath, func(relPath RelPath) error {
		visitedRelPathshMu.Lock()
		visitedRelPaths[relPath] = true
		visitedRelPathshMu.Unlock()
		if relPath == NewRelPath("dir/subdir2/file") {
			return walkErr
		}
		return nil
	}))

	// Assert that all of dir/subdir2/file's parents were visited.
	assert.True(t, visitedRelPaths[NewRelPath("dir")])
	assert.True(t, visitedRelPaths[NewRelPath("dir/subdir2")])
}
