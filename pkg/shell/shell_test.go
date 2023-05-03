package shell

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestCurrentUserShell(t *testing.T) {
	shell, ok := CurrentUserShell()
	assert.True(t, ok)
	assert.NotEqual(t, "", shell)
}
