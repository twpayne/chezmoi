package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func init() {
	// github.com/twpayne/chezmoi/v2/pkg/chezmoi reads the umask before
	// github.com/twpayne/chezmoi/v2/pkg/chezmoitest sets it, so update it.
	chezmoi.Umask = chezmoitest.Umask
}

func TestMustGetLongHelpPanics(t *testing.T) {
	assert.Panics(t, func() {
		mustLongHelp("non-existent-command")
	})
}
