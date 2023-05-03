//go:build !windows

package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestUmask(t *testing.T) {
	assert.Equal(t, chezmoitest.Umask, Umask)
}
