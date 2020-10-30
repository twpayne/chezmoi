// +build !windows

package cmd

import (
	"io"
)

// enableVirtualTerminalProcessingOnWindows does nothing on POSIX systems.
func enableVirtualTerminalProcessingOnWindows(w io.Writer) error {
	return nil
}

func trimExecutableSuffix(s string) string {
	return s
}
