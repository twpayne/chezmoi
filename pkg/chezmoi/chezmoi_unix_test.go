//go:build !windows

package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestUmask(t *testing.T) {
	require.Equal(t, chezmoitest.Umask, Umask)
}
