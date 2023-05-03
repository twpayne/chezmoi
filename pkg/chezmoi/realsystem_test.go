package chezmoi

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/alecthomas/assert/v2"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

var _ System = &RealSystem{}

func TestRealSystemGlob(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user": map[string]any{
			"bar":            "",
			"baz":            "",
			"foo":            "",
			"dir/bar":        "",
			"dir/foo":        "",
			"dir/subdir/foo": "",
		},
	}, func(fileSystem vfs.FS) {
		system := NewRealSystem(fileSystem)
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
				actualMatches, err := system.Glob(tc.pattern)
				assert.NoError(t, err)
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
