package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestIncludeMaskSet(t *testing.T) {
	for _, tc := range []struct {
		s           string
		expected    *EntryTypeSet
		expectedErr bool
	}{
		{
			s:        "",
			expected: NewEntryTypeSet(EntryTypesNone),
		},
		{
			s:        "none",
			expected: NewEntryTypeSet(EntryTypesNone),
		},
		{
			s:        "dirs,files",
			expected: NewEntryTypeSet(EntryTypeDirs | EntryTypeFiles),
		},
		{
			s:        "all",
			expected: NewEntryTypeSet(EntryTypesAll),
		},
		{
			s:        "all,noscripts",
			expected: NewEntryTypeSet(EntryTypesAll &^ EntryTypeScripts),
		},
		{
			s:        "noscripts",
			expected: NewEntryTypeSet(EntryTypesAll &^ EntryTypeScripts),
		},
		{
			s:        "noscripts,nosymlinks",
			expected: NewEntryTypeSet(EntryTypesAll &^ (EntryTypeScripts | EntryTypeSymlinks)),
		},
		{
			s:        "symlinks,,",
			expected: NewEntryTypeSet(EntryTypeSymlinks),
		},
		{
			s:           "devices",
			expectedErr: true,
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			actual := NewEntryTypeSet(EntryTypesNone)
			err := actual.Set(tc.s)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestIncludeMaskStringSlice(t *testing.T) {
	for _, tc := range []struct {
		bits     EntryTypeBits
		expected string
	}{
		{
			bits:     EntryTypesAll,
			expected: "all",
		},
		{
			bits:     EntryTypeDirs,
			expected: "dirs",
		},
		{
			bits:     EntryTypeFiles,
			expected: "files",
		},
		{
			bits:     EntryTypeRemove,
			expected: "remove",
		},
		{
			bits:     EntryTypeScripts,
			expected: "scripts",
		},
		{
			bits:     EntryTypeSymlinks,
			expected: "symlinks",
		},
		{
			bits:     EntryTypeEncrypted,
			expected: "encrypted",
		},
		{
			bits:     EntryTypeExternals,
			expected: "externals",
		},
		{
			bits:     EntryTypesNone,
			expected: "none",
		},
		{
			bits:     EntryTypeDirs | EntryTypeFiles,
			expected: "dirs,files",
		},
	} {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, NewEntryTypeSet(tc.bits).String())
		})
	}
}

func TestEntryTypeSetFlagCompletionFunc(t *testing.T) {
	for _, tc := range []struct {
		toComplete          string
		expectedCompletions []string
	}{
		{
			toComplete: "a",
			expectedCompletions: []string{
				"all",
				"always",
			},
		},
		{
			toComplete: "e",
			expectedCompletions: []string{
				"encrypted",
				"externals",
			},
		},
		{
			toComplete: "t",
			expectedCompletions: []string{
				"templates",
			},
		},
		{
			toComplete: "all,nos",
			expectedCompletions: []string{
				"all,noscripts",
				"all,nosymlinks",
			},
		},
	} {
		t.Run(tc.toComplete, func(t *testing.T) {
			completions, shellCompDirective := EntryTypeSetFlagCompletionFunc(nil, nil, tc.toComplete)
			assert.Equal(t, tc.expectedCompletions, completions)
			assert.Equal(t, cobra.ShellCompDirectiveNoSpace|cobra.ShellCompDirectiveNoFileComp, shellCompDirective)
		})
	}
}
