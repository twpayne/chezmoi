//go:build unix

package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/v2/internal/chezmoitest"
)

func TestUmask(t *testing.T) {
	assert.Equal(t, chezmoitest.Umask, Umask)
}
