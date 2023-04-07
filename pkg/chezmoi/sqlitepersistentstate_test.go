package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

var _ PersistentState = &SQLitePersistentState{}

func TestSQLitePersistentState(t *testing.T) {
	testPersistentState(t, func() PersistentState {
		s, err := NewSQLitePersistentState(":memory:")
		assert.NoError(t, err)
		return s
	})
}
