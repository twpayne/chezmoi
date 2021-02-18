package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			expected: NewEntryTypeSet(EntryTypeDirs | EntryTypeFiles | EntryTypeRemove | EntryTypeSymlinks | EntryTypeEncrypted),
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
