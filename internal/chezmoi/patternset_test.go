package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	vfs "github.com/twpayne/go-vfs/v5"

	"chezmoi.io/chezmoi/internal/chezmoitest"
)

func TestPatternSet(t *testing.T) {
	for _, tc := range []struct {
		name          string
		ps            *PatternSet
		expectMatches map[string]PatternSetMatchType
	}{
		{
			name: "empty",
			ps:   NewPatternSet(),
			expectMatches: map[string]PatternSetMatchType{
				"foo": PatternSetMatchUnknown,
			},
		},
		{
			name: "exact",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"foo": PatternSetInclude,
			}),
			expectMatches: map[string]PatternSetMatchType{
				"foo": PatternSetMatchInclude,
				"bar": PatternSetMatchExclude,
			},
		},
		{
			name: "wildcard",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"b*": PatternSetInclude,
			}),
			expectMatches: map[string]PatternSetMatchType{
				"foo": PatternSetMatchExclude,
				"bar": PatternSetMatchInclude,
				"baz": PatternSetMatchInclude,
			},
		},
		{
			name: "exclude",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"b*":  PatternSetInclude,
				"baz": PatternSetExclude,
			}),
			expectMatches: map[string]PatternSetMatchType{
				"foo": PatternSetMatchUnknown,
				"bar": PatternSetMatchInclude,
				"baz": PatternSetMatchExclude,
			},
		},
		{
			name: "doublestar",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"**/foo": PatternSetInclude,
			}),
			expectMatches: map[string]PatternSetMatchType{
				"foo":         PatternSetMatchInclude,
				"bar/foo":     PatternSetMatchInclude,
				"baz/bar/foo": PatternSetMatchInclude,
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

func TestPatternSetGlob(t *testing.T) {
	for _, tc := range []struct {
		name            string
		ps              *PatternSet
		root            any
		expectedMatches []string
	}{
		{
			name: "empty",
			ps:   NewPatternSet(),
		},
		{
			name: "simple",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"/f*": PatternSetInclude,
			}),
			root: map[string]any{
				"foo": "",
			},
			expectedMatches: []string{
				"foo",
			},
		},
		{
			name: "include_exclude",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"/b*": PatternSetInclude,
				"/*z": PatternSetExclude,
			}),
			root: map[string]any{
				"bar": "",
				"baz": "",
			},
			expectedMatches: []string{
				"bar",
			},
		},
		{
			name: "doublestar",
			ps: mustNewPatternSet(t, map[string]PatternSetIncludeType{
				"/**/f*": PatternSetInclude,
			}),
			root: map[string]any{
				"dir1/dir2/foo": "",
			},
			expectedMatches: []string{
				"dir1/dir2/foo",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				actualMatches, err := tc.ps.Glob(fileSystem, "/")
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMatches, actualMatches)
			})
		})
	}
}

func mustNewPatternSet(t *testing.T, patterns map[string]PatternSetIncludeType) *PatternSet {
	t.Helper()
	ps := NewPatternSet()
	for pattern, include := range patterns {
		assert.NoError(t, ps.Add(pattern, include))
	}
	return ps
}
