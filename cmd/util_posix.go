// +build !windows

package cmd

import (
	"io"
	"os"

	"github.com/google/renameio"
)

// enableVirtualTerminalProcessingOnWindows does nothing on POSIX systems.
func enableVirtualTerminalProcessingOnWindows(w io.Writer) error {
	return nil
}

func trimExecutableSuffix(s string) string {
	return s
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	return renameio.WriteFile(filename, data, perm)
}
