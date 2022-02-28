package cmd

import (
	"io"
	"io/fs"
	"os"

	"golang.org/x/sys/windows"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

const defaultEditor = "notepad.exe"

var defaultInterpreters = map[string]*chezmoi.Interpreter{
	"bat": {},
	"cmd": {},
	"com": {},
	"exe": {},
	"pl": {
		Command: "perl",
	},
	"ps1": {
		Command: "powershell",
		Args:    []string{"-NoLogo"},
	},
	"py": {
		Command: "python3",
	},
	"rb": {
		Command: "ruby",
	},
}

// enableVirtualTerminalProcessing enables virtual terminal processing. See
// https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences.
func enableVirtualTerminalProcessing(w io.Writer) error {
	file, ok := w.(*os.File)
	if !ok {
		return nil
	}
	var dwMode uint32
	if err := windows.GetConsoleMode(windows.Handle(file.Fd()), &dwMode); err != nil {
		return nil // Ignore error in the case that fd is not a terminal.
	}
	return windows.SetConsoleMode(windows.Handle(file.Fd()), dwMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func fileInfoUID(fs.FileInfo) int {
	return 0
}
