package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentUserShell(t *testing.T) {
	shell, ok := CurrentUserShell()
	assert.True(t, ok)
	assert.NotEmpty(t, shell)
}
