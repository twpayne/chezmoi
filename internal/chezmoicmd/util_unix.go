// +build !windows

package chezmoicmd

import (
	"io"
)

const defaultEditor = "vi"

// enableVirtualTerminalProcessing does nothing.
func enableVirtualTerminalProcessing(w io.Writer) error {
	return nil
}
