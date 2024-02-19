package chezmoi

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSourceStateEntryTreeNodeEmpty(t *testing.T) {
	n := newSourceStateTreeNode()
	assert.Equal(t, nil, n.get(EmptyRelPath))
	assert.Equal(t, []*sourceStateEntryTreeNode{n}, n.getNodes(EmptyRelPath))
	assert.NoError(t, n.forEach(EmptyRelPath, func(RelPath, SourceStateEntry) error {
		return errors.New("should not be called")
	}))
}

func TestSourceStateEntryTreeNodeSingle(t *testing.T) {
	n := newSourceStateTreeNode()
	sourceStateFile := &SourceStateFile{}
	n.set(NewRelPath("file"), sourceStateFile)
	assert.Equal(t, sourceStateFile, n.get(NewRelPath("file")).(*SourceStateFile))
	err := n.forEach(EmptyRelPath, func(targetRelPath RelPath, sourceStateEntry SourceStateEntry) error {
		assert.Equal(t, NewRelPath("file"), targetRelPath)
		assert.Equal(t, sourceStateFile, sourceStateEntry.(*SourceStateFile))
		return nil
	})
	assert.NoError(t, err)
}

func TestSourceStateEntryTreeNodeMultiple(t *testing.T) {
	entries := map[RelPath]SourceStateEntry{
		NewRelPath("a_file"):     &SourceStateFile{},
		NewRelPath("b_file"):     &SourceStateFile{},
		NewRelPath("c_file"):     &SourceStateFile{},
		NewRelPath("dir"):        &SourceStateDir{},
		NewRelPath("dir/a_file"): &SourceStateFile{},
		NewRelPath("dir/b_file"): &SourceStateFile{},
	}
	n := newSourceStateTreeNode()
	for targetRelPath, sourceStateEntry := range entries {
		n.set(targetRelPath, sourceStateEntry)
	}

	var targetRelPaths []RelPath
	err := n.forEach(EmptyRelPath, func(targetRelPath RelPath, sourceStateEntry SourceStateEntry) error {
		assert.Equal(t, entries[targetRelPath], sourceStateEntry)
		targetRelPaths = append(targetRelPaths, targetRelPath)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, []RelPath{
		NewRelPath("a_file"),
		NewRelPath("b_file"),
		NewRelPath("c_file"),
		NewRelPath("dir"),
		NewRelPath("dir/a_file"),
		NewRelPath("dir/b_file"),
	}, targetRelPaths)

	assert.Equal(t, entries, n.getMap())
}
