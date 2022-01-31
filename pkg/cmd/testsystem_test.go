package cmd

import (
	"os/exec"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

// A testSystem passes method calls to the underlying wrapped system, except for
// explicitly-overridden methods.
type testSystem struct {
	chezmoi.System
	outputFunc func(cmd *exec.Cmd) ([]byte, error)
}

func (s *testSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	if s.outputFunc != nil {
		return s.outputFunc(cmd)
	}
	return s.System.IdempotentCmdOutput(cmd)
}
