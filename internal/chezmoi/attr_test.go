package chezmoi

import (
	"io/fs"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/muesli/combinator"
)

func TestDirAttr(t *testing.T) {
	var dirAttrs []DirAttr
	targetNames := []string{
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
	}
	assert.NoError(t, combinator.Generate(&dirAttrs, struct {
		TargetName []string
		Exact      []bool
		External   []bool
		Private    []bool
		ReadOnly   []bool
		Remove     []bool
	}{
		TargetName: targetNames,
		Exact:      []bool{false, true},
		External:   []bool{false, true},
		Private:    []bool{false, true},
		ReadOnly:   []bool{false, true},
		Remove:     []bool{false, true},
	}))
	for _, dirAttr := range dirAttrs {
		actualSourceName := dirAttr.SourceName()
		actualDirAttr := parseDirAttr(actualSourceName)
		assert.Equal(t, dirAttr, actualDirAttr)
		assert.Equal(t, actualSourceName, actualDirAttr.SourceName())
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
	var fileAttrs []FileAttr
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
		"remove_",
		"run_name",
		"symlink_name",
		"template.tmpl",
	}
	assert.NoError(t, combinator.Generate(&fileAttrs, struct {
		TargetName []string
		Empty      []bool
		Encrypted  []bool
		Executable []bool
		Private    []bool
		ReadOnly   []bool
		Template   []bool
		Type       SourceFileTargetType
	}{
		Type:       SourceFileTypeCreate,
		TargetName: []string{},
		Empty:      []bool{false, true},
		Encrypted:  []bool{false, true},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		ReadOnly:   []bool{false, true},
		Template:   []bool{false, true},
	}))
	assert.NoError(t, combinator.Generate(&fileAttrs, struct {
		TargetName []string
		Empty      []bool
		Encrypted  []bool
		Executable []bool
		Private    []bool
		ReadOnly   []bool
		Template   []bool
		Type       SourceFileTargetType
	}{
		Type:       SourceFileTypeFile,
		TargetName: targetNames,
		Empty:      []bool{false, true},
		Encrypted:  []bool{false, true},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		ReadOnly:   []bool{false, true},
		Template:   []bool{false, true},
	}))
	assert.NoError(t, combinator.Generate(&fileAttrs, struct {
		TargetName []string
		Encrypted  []bool
		Executable []bool
		Private    []bool
		ReadOnly   []bool
		Template   []bool
		Type       SourceFileTargetType
	}{
		Type:       SourceFileTypeModify,
		TargetName: targetNames,
		Encrypted:  []bool{false, true},
		Executable: []bool{false, true},
		Private:    []bool{false, true},
		ReadOnly:   []bool{false, true},
		Template:   []bool{false, true},
	}))
	assert.NoError(t, combinator.Generate(&fileAttrs, struct {
		TargetName []string
		Type       SourceFileTargetType
	}{
		Type:       SourceFileTypeRemove,
		TargetName: targetNames,
	}))
	assert.NoError(t, combinator.Generate(&fileAttrs, struct {
		Condition  []ScriptCondition
		TargetName []string
		Order      []ScriptOrder
		Type       SourceFileTargetType
	}{
		Type: SourceFileTypeScript,
		Condition: []ScriptCondition{
			ScriptConditionAlways,
			ScriptConditionOnce,
			ScriptConditionOnChange,
		},
		TargetName: targetNames,
		Order:      []ScriptOrder{ScriptOrderBefore, ScriptOrderDuring, ScriptOrderAfter},
	}))
	assert.NoError(t, combinator.Generate(&fileAttrs, struct {
		TargetName []string
		Type       SourceFileTargetType
	}{
		Type:       SourceFileTypeSymlink,
		TargetName: targetNames,
	}))
	for _, fileAttr := range fileAttrs {
		actualSourceName := fileAttr.SourceName("")
		actualFileAttr := parseFileAttr(actualSourceName, "")
		assert.Equal(t, fileAttr, actualFileAttr)
		assert.Equal(t, actualSourceName, actualFileAttr.SourceName(""))
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
				Condition:  ScriptConditionOnce,
				Type:       SourceFileTypeScript,
			},
		},
		{
			sourceName: "run_literal_once_script",
			fileAttr: FileAttr{
				TargetName: "once_script",
				Condition:  ScriptConditionAlways,
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

func TestFileAttrPerm(t *testing.T) {
	for _, tc := range []struct {
		fileAttr FileAttr
		expected fs.FileMode
	}{
		{
			fileAttr: FileAttr{},
			expected: 0o666,
		},
		{
			fileAttr: FileAttr{
				Executable: true,
			},
			expected: fs.ModePerm,
		},
		{
			fileAttr: FileAttr{
				Private: true,
			},
			expected: 0o600,
		},
		{
			fileAttr: FileAttr{
				Executable: true,
				Private:    true,
			},
			expected: 0o700,
		},
		{
			fileAttr: FileAttr{
				ReadOnly: true,
			},
			expected: 0o444,
		},
		{
			fileAttr: FileAttr{
				Executable: true,
				ReadOnly:   true,
			},
			expected: 0o555,
		},
		{
			fileAttr: FileAttr{
				Private:  true,
				ReadOnly: true,
			},
			expected: 0o400,
		},
		{
			fileAttr: FileAttr{
				Executable: true,
				Private:    true,
				ReadOnly:   true,
			},
			expected: 0o500,
		},
	} {
		assert.Equal(t, tc.expected, tc.fileAttr.perm())
	}
}
