package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFS(t *testing.T) {
	_, err := FS.ReadFile("COMMIT_MESSAGE.tmpl")
	require.NoError(t, err)
}
