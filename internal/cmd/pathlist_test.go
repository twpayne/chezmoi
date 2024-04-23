package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestNewNodeFromPathsSlice(t *testing.T) {
	for _, tc := range []struct {
		name     string
		paths    []string
		expected []string
	}{
		{
			name: "empty",
		},
		{
			name: "root",
			paths: []string{
				"a",
			},
			expected: []string{
				"a",
			},
		},
		{
			name: "simple",
			paths: []string{
				"a",
				"b",
			},
			expected: []string{
				"a",
				"b",
			},
		},
		{
			name: "simple_nesting",
			paths: []string{
				"a/b",
			},
			expected: []string{
				"a",
				"  b",
			},
		},
		{
			name: "multiple_simple_nesting",
			paths: []string{
				"a/a",
				"a/b",
				"b/a",
				"b/b",
			},
			expected: []string{
				"a",
				"  a",
				"  b",
				"b",
				"  a",
				"  b",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sb strings.Builder
			newPathListTreeFromPathsSlice(tc.paths).writeChildren(&sb, "", "  ")
			assert.Equal(t, chezmoitest.JoinLines(tc.expected...), sb.String())
		})
	}
}
