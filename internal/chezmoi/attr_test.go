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
			"exact_dir",
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

func TestDirAttrLiteral(t *testing.T) {
	for _, tc := range []struct {
		sourceName string
		dirAttr    DirAttr
	}{
		{
			sourceName: "exact_dir",
			dirAttr: DirAttr{
				TargetName: "dir",
				Exact:      true,
			},
		},
		{
			sourceName: "literal_exact_dir",
			dirAttr: DirAttr{
				TargetName: "exact_dir",
			},
		},
		{
			sourceName: "literal_literal_dir",
			dirAttr: DirAttr{
				TargetName: "literal_dir",
			},
		},
	} {
		t.Run(tc.sourceName, func(t *testing.T) {
			assert.Equal(t, tc.sourceName, tc.dirAttr.SourceName())
			assert.Equal(t, tc.dirAttr, parseDirAttr(tc.sourceName))
		})
	}
}

func TestFileAttr(t *testing.T) {
	var fas []FileAttr
	targetNames := []string{
		".name",
		"create_name",
		"dot_name",
		"exact_name",
		"literal_name",
		"literal_name",
		"modify_name",
		"name.literal",
		"name",
		"run_name",
		"symlink_name",
		"template.tmpl",
	}
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
		Encrypted  []bool
		Executable []bool
		Private    []bool
		Template   []bool
	}{
		Type:       SourceFileTypeCreate,
		TargetName: []string{},
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
		Type:       SourceFileTypeFile,
		TargetName: targetNames,
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
		Type:       SourceFileTypeModify,
		TargetName: targetNames,
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		Template:   []bool{false, true},
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
	}{
		Type:       SourceFileTypeRemove,
		TargetName: targetNames,
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
		Once       []bool
		Order      []int
	}{
		Type:       SourceFileTypeScript,
		TargetName: targetNames,
		Once:       []bool{false, true},
		Order:      []int{-1, 0, 1},
	}))
	require.NoError(t, combinator.Generate(&fas, struct {
		Type       SourceFileTargetType
		TargetName []string
	}{
		Type:       SourceFileTypeSymlink,
		TargetName: targetNames,
	}))
	for _, fa := range fas {
		actualSourceName := fa.SourceName("")
		actualFA := parseFileAttr(actualSourceName, "")
		assert.Equal(t, fa, actualFA)
		assert.Equal(t, actualSourceName, actualFA.SourceName(""))
	}
}

func TestFileAttrEncryptedSuffix(t *testing.T) {
	for _, tc := range []struct {
		sourceName         string
		expectedTargetName string
	}{
		{
			sourceName:         "encrypted_file",
			expectedTargetName: "file",
		},
		{
			sourceName:         "encrypted_file.asc",
			expectedTargetName: "file",
		},
		{
			sourceName:         "file.asc",
			expectedTargetName: "file.asc",
		},
	} {
		fa := parseFileAttr(tc.sourceName, ".asc")
		assert.Equal(t, tc.expectedTargetName, fa.TargetName)
	}
}

func TestFileAttrLiteral(t *testing.T) {
	for _, tc := range []struct {
		sourceName      string
		encryptedSuffix string
		fileAttr        FileAttr
		nonCanonical    bool
	}{
		{
			sourceName: "dot_file",
			fileAttr: FileAttr{
				TargetName: ".file",
				Type:       SourceFileTypeFile,
			},
		},
		{
			sourceName: "literal_dot_file",
			fileAttr: FileAttr{
				TargetName: "dot_file",
				Type:       SourceFileTypeFile,
			},
		},
		{
			sourceName: "literal_literal_file",
			fileAttr: FileAttr{
				TargetName: "literal_file",
				Type:       SourceFileTypeFile,
			},
		},
		{
			sourceName: "run_once_script",
			fileAttr: FileAttr{
				TargetName: "script",
				Type:       SourceFileTypeScript,
				Once:       true,
			},
		},
		{
			sourceName: "run_literal_once_script",
			fileAttr: FileAttr{
				TargetName: "once_script",
				Type:       SourceFileTypeScript,
			},
		},
		{
			sourceName: "file.literal",
			fileAttr: FileAttr{
				TargetName: "file",
				Type:       SourceFileTypeFile,
			},
			nonCanonical: true,
		},
		{
			sourceName: "file.literal.literal",
			fileAttr: FileAttr{
				TargetName: "file.literal",
				Type:       SourceFileTypeFile,
			},
		},
		{
			sourceName: "file.tmpl",
			fileAttr: FileAttr{
				TargetName: "file",
				Type:       SourceFileTypeFile,
				Template:   true,
			},
		},
		{
			sourceName: "file.tmpl.literal",
			fileAttr: FileAttr{
				TargetName: "file.tmpl",
				Type:       SourceFileTypeFile,
			},
		},
		{
			sourceName: "file.tmpl.literal.tmpl",
			fileAttr: FileAttr{
				TargetName: "file.tmpl",
				Type:       SourceFileTypeFile,
				Template:   true,
			},
		},
	} {
		t.Run(tc.sourceName, func(t *testing.T) {
			assert.Equal(t, tc.fileAttr, parseFileAttr(tc.sourceName, tc.encryptedSuffix))
			if !tc.nonCanonical {
				assert.Equal(t, tc.sourceName, tc.fileAttr.SourceName(tc.encryptedSuffix))
			}
		})
	}
}
