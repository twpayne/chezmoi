// +build !windows

package cmd

import (
	"io"
)

// enableVirtualTerminalProcessing does nothing.
func enableVirtualTerminalProcessing(w io.Writer) error {
	return nil
}
