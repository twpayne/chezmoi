package chezmoi

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/muesli/combinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntryStateEquivalent(t *testing.T) {
	entryStates := map[string]*EntryState{
		"absent": {
			Type: EntryStateTypeAbsent,
		},
		"dir1": {
			Type: EntryStateTypeDir,
			Mode: os.ModeDir | 0o777,
		},
		"dir1_copy": {
			Type: EntryStateTypeDir,
			Mode: os.ModeDir | 0o777,
		},
		"dir_private": {
			Type: EntryStateTypeDir,
			Mode: os.ModeDir | 0o700,
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
		"present": {
			Type:           EntryStateTypePresent,
			Mode:           0o666,
			ContentsSHA256: []byte{3},
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
		"absent_nil1":          true,
		"absent_nil2":          true,
		"dir1_copy_dir1":       true,
		"dir1_dir1_copy":       true,
		"file1_copy_file1":     true,
		"file1_copy_present":   true,
		"file1_file1_copy":     true,
		"file1_present":        true,
		"file2_present":        true,
		"nil1_absent":          true,
		"nil2_absent":          true,
		"present_file1_copy":   true,
		"present_file1":        true,
		"present_file2":        true,
		"symlink_copy_symlink": true,
		"symlink_symlink_copy": true,
	}

	entryStateKeys := make([]string, 0, len(entryStates))
	for entryStateKey := range entryStates {
		entryStateKeys = append(entryStateKeys, entryStateKey)
	}
	sort.Strings(entryStateKeys)

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
	require.NoError(t, combinator.Generate(&testCases, testData))

	for _, tc := range testCases {
		name := fmt.Sprintf("%s_%s", tc.EntryState1Key, tc.EntryState2Key)
		t.Run(name, func(t *testing.T) {
			entryState1 := entryStates[tc.EntryState1Key]
			entryState2 := entryStates[tc.EntryState2Key]
			expectedEquivalent := entryState1 == entryState2 || expectedEquivalents[name]
			assert.Equal(t, expectedEquivalent, entryState1.Equivalent(entryState2, 0o022))
		})
	}
}
