package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncludeMaskSet(t *testing.T) {
	for _, tc := range []struct {
		s           string
		expected    *IncludeSet
		expectedErr bool
	}{
		{
			s:        "",
			expected: NewIncludeSet(includeNone),
		},
		{
			s:        "none",
			expected: NewIncludeSet(includeNone),
		},
		{
			s:        "dirs,files",
			expected: NewIncludeSet(IncludeDirs | IncludeFiles),
		},
		{
			s:        "all",
			expected: NewIncludeSet(IncludeAll),
		},
		{
			s:        "all,!scripts",
			expected: NewIncludeSet(IncludeAbsent | IncludeDirs | IncludeFiles | IncludeSymlinks),
		},
		{
			s:        "a,s",
			expected: NewIncludeSet(IncludeAbsent | IncludeSymlinks),
		},
		{
			s:        "symlinks,,",
			expected: NewIncludeSet(IncludeSymlinks),
		},
		{
			s:           "devices",
			expectedErr: true,
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			actual := NewIncludeSet(includeNone)
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
		bits     IncludeBits
		expected string
	}{
		{
			bits:     IncludeAll,
			expected: "all",
		},
		{
			bits:     IncludeAbsent,
			expected: "absent",
		},
		{
			bits:     IncludeDirs,
			expected: "dirs",
		},
		{
			bits:     IncludeFiles,
			expected: "files",
		},
		{
			bits:     IncludeScripts,
			expected: "scripts",
		},
		{
			bits:     IncludeSymlinks,
			expected: "symlinks",
		},
		{
			bits:     includeNone,
			expected: "none",
		},
		{
			bits:     IncludeDirs | IncludeFiles,
			expected: "dirs,files",
		},
	} {
		assert.Equal(t, tc.expected, NewIncludeSet(tc.bits).String())
	}
}
