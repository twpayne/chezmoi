package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFS(t *testing.T) {
	_, err := FS.ReadFile("add.md")
	require.NoError(t, err)
}
