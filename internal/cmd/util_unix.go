// +build !windows

package cmd

import (
	"io"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

const defaultEditor = "vi"

var defaultInterpreters = make(map[string]*chezmoi.Interpreter)

// enableVirtualTerminalProcessing does nothing.
func enableVirtualTerminalProcessing(w io.Writer) error {
	return nil
}
