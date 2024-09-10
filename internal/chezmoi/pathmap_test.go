package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPathMap(t *testing.T) {
	pm := NewPathMap()
	assert.NoError(t, pm.AddStringMap(map[string]string{
		".config/glow":       "Library/Preferences/glow",
		".local/share/fonts": "Library/Fonts",
	}))

	assert.Equal(t, NewRelPath(".config"), pm.Lookup(NewRelPath(".config")))
	assert.Equal(t, NewRelPath("Library/Preferences/glow"), pm.Lookup(NewRelPath(".config/glow")))
	assert.Equal(t, NewRelPath(".config/nvim"), pm.Lookup(NewRelPath(".config/nvim")))
	assert.Equal(t, NewRelPath(".local"), pm.Lookup(NewRelPath(".local")))
	assert.Equal(t, NewRelPath(".local/share"), pm.Lookup(NewRelPath(".local/share")))
	assert.Equal(t, NewRelPath(".local/share/chezmoi"), pm.Lookup(NewRelPath(".local/share/chezmoi")))
	assert.Equal(t, NewRelPath("Library/Fonts"), pm.Lookup(NewRelPath(".local/share/fonts")))
}

func TestPathMapErrors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pairs       [][2]string
		expectedErr string
	}{
		{
			name: "empty",
		},
		{
			name: "parent_and_child",
			pairs: [][2]string{
				{"a", "x"},
				{"a/b", "y"},
			},
			expectedErr: "a/b -> y: parent a is already mapped to x",
		},
		{
			name: "child_and_parent",
			pairs: [][2]string{
				{"a/b", "x"},
				{"a", "y"},
			},
			expectedErr: "a -> y: directory is already mapped",
		},
		{
			name: "duplicate_inconsistent",
			pairs: [][2]string{
				{"a/b", "x"},
				{"a/b", "y"},
			},
			expectedErr: "a/b -> y: directory is already mapped",
		},
		{
			name: "duplicate_consistent",
			pairs: [][2]string{
				{"a/b", "x"},
				{"a/b", "x"},
			},
			expectedErr: "a/b -> x: directory is already mapped",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pm := NewPathMap()
			for i, pair := range tc.pairs {
				err := pm.Add(NewRelPath(pair[0]), NewRelPath(pair[1]))
				if i != len(tc.pairs)-1 || tc.expectedErr == "" {
					assert.NoError(t, err)
				} else {
					assert.EqualError(t, err, tc.expectedErr)
				}
			}
		})
	}
}
