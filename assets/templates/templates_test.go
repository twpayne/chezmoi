package templates

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFS(t *testing.T) {
	_, err := FS.ReadFile("COMMIT_MESSAGE.tmpl")
	assert.NoError(t, err)
}
