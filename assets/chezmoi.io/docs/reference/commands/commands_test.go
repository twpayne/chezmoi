package commands

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFS(t *testing.T) {
	_, err := FS.ReadFile("add.md")
	assert.NoError(t, err)
}
