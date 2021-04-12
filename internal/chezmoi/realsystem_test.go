package chezmoi

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

var _ System = &RealSystem{}

func TestRealSystemGlob(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]interface{}{
		"/home/user": map[string]interface{}{
			"bar":            "",
			"baz":            "",
			"foo":            "",
			"dir/bar":        "",
			"dir/foo":        "",
			"dir/subdir/foo": "",
		},
	}, func(fs vfs.FS) {
		s := NewRealSystem(fs)
		for _, tc := range []struct {
			pattern         string
			expectedMatches []string
		}{
			{
				pattern: "/home/user/foo",
				expectedMatches: []string{
					"/home/user/foo",
				},
			},
			{
				pattern: "/home/user/**/foo",
				expectedMatches: []string{
					"/home/user/dir/foo",
					"/home/user/dir/subdir/foo",
					"/home/user/foo",
				},
			},
			{
				pattern: "/home/user/**/ba*",
				expectedMatches: []string{
					"/home/user/bar",
					"/home/user/baz",
					"/home/user/dir/bar",
				},
			},
		} {
			t.Run(tc.pattern, func(t *testing.T) {
				actualMatches, err := s.Glob(tc.pattern)
				require.NoError(t, err)
				sort.Strings(actualMatches)
				assert.Equal(t, tc.expectedMatches, pathsToSlashes(actualMatches))
			})
		}
	})
}

func pathsToSlashes(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, filepath.ToSlash(path))
	}
	return result
}
