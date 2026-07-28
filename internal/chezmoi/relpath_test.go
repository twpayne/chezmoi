package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestNewUntrustedRelPath(t *testing.T) {
	for p, ok := range map[string]bool{
		"":       false,
		"..":     false,
		"a":      true,
		"a/b":    true,
		"../a":   false,
		"a/../b": false,
		"a/..":   false,
	} {
		t.Run(p, func(t *testing.T) {
			_, err := NewUntrustedRelPath(p)
			if ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
