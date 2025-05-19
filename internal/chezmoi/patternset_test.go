package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	vfs "github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestPatternSet(t *testing.T) {
	for _, tc := range []struct {
		name          string
		ps            *patternSet
		expectMatches map[string]patternSetMatchType
	}{
		{
			name: "empty",
			ps:   newPatternSet(),
			expectMatches: map[string]patternSetMatchType{
				"foo": patternSetMatchUnknown,
			},
		},
		{
			name: "exact",
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"foo": patternSetInclude,
			}),
			expectMatches: map[string]patternSetMatchType{
				"foo": patternSetMatchInclude,
				"bar": patternSetMatchExclude,
			},
		},
		{
			name: "wildcard",
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"b*": patternSetInclude,
			}),
			expectMatches: map[string]patternSetMatchType{
				"foo": patternSetMatchExclude,
				"bar": patternSetMatchInclude,
				"baz": patternSetMatchInclude,
			},
		},
		{
			name: "exclude",
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"b*":  patternSetInclude,
				"baz": patternSetExclude,
			}),
			expectMatches: map[string]patternSetMatchType{
				"foo": patternSetMatchUnknown,
				"bar": patternSetMatchInclude,
				"baz": patternSetMatchExclude,
			},
		},
		{
			name: "doublestar",
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"**/foo": patternSetInclude,
			}),
			expectMatches: map[string]patternSetMatchType{
				"foo":         patternSetMatchInclude,
				"bar/foo":     patternSetMatchInclude,
				"baz/bar/foo": patternSetMatchInclude,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for s, expectMatch := range tc.expectMatches {
				assert.Equal(t, expectMatch, tc.ps.match(s))
			}
		})
	}
}

func TestPatternSetGlob(t *testing.T) {
	for _, tc := range []struct {
		name            string
		ps              *patternSet
		root            any
		expectedMatches []string
	}{
		{
			name:            "empty",
			ps:              newPatternSet(),
			root:            nil,
			expectedMatches: []string{},
		},
		{
			name: "simple",
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"/f*": patternSetInclude,
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
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"/b*": patternSetInclude,
				"/*z": patternSetExclude,
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
			ps: mustNewPatternSet(t, map[string]patternSetIncludeType{
				"/**/f*": patternSetInclude,
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
				actualMatches, err := tc.ps.glob(fileSystem, "/")
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMatches, actualMatches)
			})
		})
	}
}

func mustNewPatternSet(t *testing.T, patterns map[string]patternSetIncludeType) *patternSet {
	t.Helper()
	ps := newPatternSet()
	for pattern, include := range patterns {
		assert.NoError(t, ps.add(pattern, include))
	}
	return ps
}
