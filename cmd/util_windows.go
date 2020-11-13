package cmd

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

// enableVirtualTerminalProcessingOnWindows enables virtual terminal processing
// on Windows. See
// https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences.
func enableVirtualTerminalProcessingOnWindows(w io.Writer) error {
	f, ok := w.(*os.File)
	if !ok {
		return nil
	}
	var dwMode uint32
	if err := windows.GetConsoleMode(windows.Handle(f.Fd()), &dwMode); err != nil {
		return nil // Ignore error in the case that fd is not a terminal.
	}
	return windows.SetConsoleMode(windows.Handle(f.Fd()), dwMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func trimExecutableSuffix(s string) string {
	return strings.TrimSuffix(s, ".exe")
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}
