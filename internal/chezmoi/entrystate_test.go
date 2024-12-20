package chezmoi

import (
	"fmt"
	"io/fs"
	"maps"
	"runtime"
	"slices"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/muesli/combinator"
)

func TestEntryStateEquivalent(t *testing.T) {
	entryStates := map[string]*EntryState{
		"dir1": {
			Type: EntryStateTypeDir,
			Mode: fs.ModeDir | fs.ModePerm,
		},
		"dir1_copy": {
			Type: EntryStateTypeDir,
			Mode: fs.ModeDir | fs.ModePerm,
		},
		"dir_private": {
			Type: EntryStateTypeDir,
			Mode: fs.ModeDir | 0o700,
		},
		"file1": {
			Type:           EntryStateTypeFile,
			Mode:           0o666,
			ContentsSHA256: []byte{1},
		},
		"file1_copy": {
			Type:           EntryStateTypeFile,
			Mode:           0o666,
			ContentsSHA256: []byte{1},
		},
		"file2": {
			Type:           EntryStateTypeFile,
			Mode:           0o666,
			ContentsSHA256: []byte{2},
		},
		"nil1": nil,
		"nil2": nil,
		"remove": {
			Type: EntryStateTypeRemove,
		},
		"script": {
			Type:           EntryStateTypeScript,
			ContentsSHA256: []byte{4},
		},
		"symlink": {
			Type:           EntryStateTypeSymlink,
			ContentsSHA256: []byte{5},
		},
		"symlink_copy": {
			Type:           EntryStateTypeSymlink,
			ContentsSHA256: []byte{5},
		},
	}

	expectedEquivalents := map[string]bool{
		"dir_private_dir1_copy": runtime.GOOS == "windows",
		"dir_private_dir1":      runtime.GOOS == "windows",
		"dir1_copy_dir_private": runtime.GOOS == "windows",
		"dir1_copy_dir1":        true,
		"dir1_dir_private":      runtime.GOOS == "windows",
		"dir1_dir1_copy":        true,
		"file1_copy_file1":      true,
		"file1_create":          true,
		"file1_file1_copy":      true,
		"nil1_remove":           true,
		"nil2_remove":           true,
		"remove_nil1":           true,
		"remove_nil2":           true,
		"symlink_copy_symlink":  true,
		"symlink_symlink_copy":  true,
	}

	entryStateKeys := slices.Sorted(maps.Keys(entryStates))

	testData := struct {
		EntryState1Key []string
		EntryState2Key []string
	}{
		EntryState1Key: entryStateKeys,
		EntryState2Key: entryStateKeys,
	}
	var testCases []struct {
		EntryState1Key string
		EntryState2Key string
	}
	assert.NoError(t, combinator.Generate(&testCases, testData))

	for _, tc := range testCases {
		name := fmt.Sprintf("%s_%s", tc.EntryState1Key, tc.EntryState2Key)
		t.Run(name, func(t *testing.T) {
			entryState1 := entryStates[tc.EntryState1Key]
			entryState2 := entryStates[tc.EntryState2Key]
			expectedEquivalent := entryState1 == entryState2 || expectedEquivalents[name]
			assert.Equal(t, expectedEquivalent, entryState1.Equivalent(entryState2))
		})
	}
}
