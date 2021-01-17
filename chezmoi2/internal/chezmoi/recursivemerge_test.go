package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecursiveMerge(t *testing.T) {
	for _, tc := range []struct {
		dest         map[string]interface{}
		source       map[string]interface{}
		expectedDest map[string]interface{}
	}{
		{
			dest:         map[string]interface{}{},
			source:       nil,
			expectedDest: map[string]interface{}{},
		},
		{
			dest: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": map[string]interface{}{
					"d": 4,
					"e": 5,
				},
				"f": map[string]interface{}{
					"g": 6,
				},
			},
			source: map[string]interface{}{
				"b": 20,
				"c": map[string]interface{}{
					"e": 50,
					"f": 60,
				},
				"f": 60,
			},
			expectedDest: map[string]interface{}{
				"a": 1,
				"b": 20,
				"c": map[string]interface{}{
					"d": 4,
					"e": 50,
					"f": 60,
				},
				"f": 60,
			},
		},
	} {
		recursiveMerge(tc.dest, tc.source)
		assert.Equal(t, tc.expectedDest, tc.dest)
	}
}

func TestRecursiveMergeCopies(t *testing.T) {
	original := map[string]interface{}{
		"key": "initialValue",
	}
	dest := make(map[string]interface{})
	recursiveMerge(dest, original)
	recursiveMerge(dest, map[string]interface{}{
		"key": "mergedValue",
	})
	assert.Equal(t, "mergedValue", dest["key"])
	assert.Equal(t, "initialValue", original["key"])
}
