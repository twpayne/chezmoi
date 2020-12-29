package chezmoi

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternSet(t *testing.T) {
	for _, tc := range []struct {
		name          string
		ps            *PatternSet
		expectMatches map[string]bool
	}{
		{
			name: "empty",
			ps:   NewPatternSet(),
			expectMatches: map[string]bool{
				"foo": false,
			},
		},
		{
			name: "exact",
			ps: mustNewPatternSet(t, map[string]bool{
				"foo": true,
			}),
			expectMatches: map[string]bool{
				"foo": true,
				"bar": false,
			},
		},
		{
			name: "wildcard",
			ps: mustNewPatternSet(t, map[string]bool{
				"b*": true,
			}),
			expectMatches: map[string]bool{
				"foo": false,
				"bar": true,
				"baz": true,
			},
		},
		{
			name: "exclude",
			ps: mustNewPatternSet(t, map[string]bool{
				"b*":  true,
				"baz": false,
			}),
			expectMatches: map[string]bool{
				"foo": false,
				"bar": true,
				"baz": false,
			},
		},
		{
			name: "doublestar",
			ps: mustNewPatternSet(t, map[string]bool{
				"**/foo": true,
			}),
			expectMatches: map[string]bool{
				"foo":                              true,
				filepath.Join("bar", "foo"):        true,
				filepath.Join("baz", "bar", "foo"): true,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for s, expectMatch := range tc.expectMatches {
				assert.Equal(t, expectMatch, tc.ps.Match(s))
			}
		})
	}
}

func mustNewPatternSet(t *testing.T, patterns map[string]bool) *PatternSet {
	t.Helper()
	ps := NewPatternSet()
	for pattern, exclude := range patterns {
		require.NoError(t, ps.Add(pattern, exclude))
	}
	return ps
}
