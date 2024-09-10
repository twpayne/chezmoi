package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBiDiPathMap(t *testing.T) {
	m := NewBiDiPathMap()
	assert.NoError(t, m.AddStringMap(map[string]string{
		".config/glow":       "Library/Preferences/glow",
		".local/share/fonts": "Library/Fonts",
	}))
	assert.Equal(t, NewRelPath("Library/Preferences/glow"), m.Forward(NewRelPath(".config/glow")))
	assert.Equal(t, NewRelPath(".config/glow"), m.Reverse(NewRelPath("Library/Preferences/glow")))
}
