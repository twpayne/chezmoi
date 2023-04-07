package chezmoi

import (
	"testing"
)

func TestMockPersistentState(t *testing.T) {
	testPersistentState(t, func() PersistentState {
		return NewMockPersistentState()
	})
}
