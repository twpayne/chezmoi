package chezmoi

import (
	"testing"

	"github.com/muesli/combinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirAttr(t *testing.T) {
	testData := struct {
		TargetName []string
		Exact      []bool
		Private    []bool
	}{
		TargetName: []string{
			".dir",
			"dir.tmpl",
			"dir",
			"empty_dir",
			"encrypted_dir",
			"executable_dir",
			"once_dir",
			"run_dir",
			"run_once_dir",
			"symlink_dir",
		},
		Exact:   []bool{false, true},
		Private: []bool{false, true},
	}
	var das []DirAttr
	require.NoError(t, combinator.Generate(&das, testData))
	for _, da := range das {
		actualSourceName := da.SourceName()
		actualDA := parseDirAttr(actualSourceName)
		assert.Equal(t, da, actualDA)
		assert.Equal(t, actualSourceName, actualDA.SourceName())
	}
}

func TestFileAttr(t *testing.T) {
	var fas []FileAttr
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
		Encrypted  []bool
		Executable []bool
		Private    []bool
		Template   []bool
	}{
		Type: SourceFileTypeCreate,
		TargetName: []string{
			".name",
			"exact_name",
			"name",
		},
		Encrypted:  []bool{false, true},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		Template:   []bool{false, true},
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
		Empty      []bool
		Encrypted  []bool
		Executable []bool
		Private    []bool
		Template   []bool
	}{
		Type: SourceFileTypeFile,
		TargetName: []string{
			".name",
			"exact_name",
			"name",
		},
		Empty:      []bool{false, true},
		Encrypted:  []bool{false, true},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		Template:   []bool{false, true},
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
		Executable []bool
		Private    []bool
		Template   []bool
	}{
		Type: SourceFileTypeModify,
		TargetName: []string{
			".name",
			"exact_name",
			"name",
		},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		Template:   []bool{false, true},
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
		Once       []bool
		Order      []int
	}{
		Type: SourceFileTypeScript,
		TargetName: []string{
			".name",
			"exact_name",
			"name",
		},
		Once:  []bool{false, true},
		Order: []int{-1, 0, 1},
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
	}{
		Type: SourceFileTypeSymlink,
		TargetName: []string{
			".name",
			"exact_name",
			"name",
		},
	}))
	for _, fa := range fas {
		actualSourceName := fa.SourceName()
		actualFA := parseFileAttr(actualSourceName)
		assert.Equal(t, fa, actualFA)
		assert.Equal(t, actualSourceName, actualFA.SourceName())
	}
}
