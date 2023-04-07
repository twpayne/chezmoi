package chezmoi

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v4"
)

var _ PersistentState = &SimplePersistentState{}

func TestSimplePersistentState(t *testing.T) {
	system := NewRealSystem(vfs.OSFS)
	var tempDirs []string
	defer func() {
		for _, tempDir := range tempDirs {
			assert.NoError(t, os.RemoveAll(tempDir))
		}
	}()
	testPersistentState(t, func() PersistentState {
		tempDir, err := os.MkdirTemp("", "chezmoi-test")
		assert.NoError(t, err)
		return NewSimplePersistentState(system, NewAbsPath(tempDir).JoinString("chezmoistate.json"))
	})
}
