package cmd

import (
	"fmt"
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

// enableVirtualTerminalProcessing enables virtual terminal processing on
// Windows systems. See
// https://docs.microsoft.com/en-us/windows/console/console-virtual-terminal-sequences.
// It returns a function that restores the console to the previous state.
func enableVirtualTerminalProcessing(w io.Writer) (func() error, error) {
	file, ok := w.(*os.File)
	if !ok {
		return nil, nil
	}

	var dwMode uint32
	if err := windows.GetConsoleMode(windows.Handle(file.Fd()), &dwMode); err != nil {
		// Ignore error in the case that fd is not a terminal.
		return nil, nil
	}

	if dwMode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING {
		// If virtual terminal processing is already enabled, then there is
		// nothing to do.
		return nil, nil
	}

	if err := windows.SetConsoleMode(windows.Handle(file.Fd()), dwMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return nil, fmt.Errorf("windows.SetConsoleMode: %w", err)
	}

	return func() error {
		return windows.SetConsoleMode(windows.Handle(file.Fd()), dwMode)
	}, nil
}

func fileInfoUID(fs.FileInfo) int {
	return 0
}
