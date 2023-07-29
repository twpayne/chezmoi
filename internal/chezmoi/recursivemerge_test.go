package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestRecursiveMerge(t *testing.T) {
	for _, tc := range []struct {
		dest         map[string]any
		source       map[string]any
		expectedDest map[string]any
	}{
		{
			dest:         map[string]any{},
			source:       nil,
			expectedDest: map[string]any{},
		},
		{
			dest: map[string]any{
				"a": 1,
				"b": 2,
				"c": map[string]any{
					"d": 4,
					"e": 5,
				},
				"f": map[string]any{
					"g": 6,
				},
			},
			source: map[string]any{
				"b": 20,
				"c": map[string]any{
					"e": 50,
					"f": 60,
				},
				"f": 60,
			},
			expectedDest: map[string]any{
				"a": 1,
				"b": 20,
				"c": map[string]any{
					"d": 4,
					"e": 50,
					"f": 60,
				},
				"f": 60,
			},
		},
	} {
		RecursiveMerge(tc.dest, tc.source)
		assert.Equal(t, tc.expectedDest, tc.dest)
	}
}

func TestRecursiveMergeCopies(t *testing.T) {
	original := map[string]any{
		"key": "initialValue",
	}
	dest := make(map[string]any)
	RecursiveMerge(dest, original)
	RecursiveMerge(dest, map[string]any{
		"key": "mergedValue",
	})
	assert.Equal(t, "mergedValue", dest["key"])
	assert.Equal(t, "initialValue", original["key"])
}
