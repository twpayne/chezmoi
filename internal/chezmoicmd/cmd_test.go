package chezmoicmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func init() {
	// github.com/twpayne/chezmoi/v2/internal/chezmoi reads the umask before
	// github.com/twpayne/chezmoi/v2/internal/chezmoitest sets it, so update it.
	chezmoi.Umask = chezmoitest.Umask
}

func TestMustGetLongHelpPanics(t *testing.T) {
	assert.Panics(t, func() {
		mustLongHelp("non-existent-command")
	})
}
