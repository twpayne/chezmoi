//go:build unix

package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/chezmoitest"
)

func TestUmask(t *testing.T) {
	assert.Equal(t, chezmoitest.Umask, Umask)
}
