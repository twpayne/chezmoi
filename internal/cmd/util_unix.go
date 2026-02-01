//go:build unix

package cmd

import (
	"io/fs"
	"os"
	"strings"
	"syscall"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

const defaultEditor = "vi"

func fileInfoUID(info fs.FileInfo) int {
	return int(info.Sys().(*syscall.Stat_t).Uid) //nolint:forcetypeassert
}

// getPS1Interpreter returns the appropriate Interpreter for PowerShell
// scripts (.ps1) on Unix-like systems. It uses the provided findExecutable
// function to check for the presence of 'pwsh'. If 'pwsh' is not found, it
// returns an empty Interpreter, indicating no suitable interpreter is
// available.
func getPS1Interpreter(findExecutable func([]string, []string) (string, error)) chezmoi.Interpreter {
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	if pwshPath, _ := findExecutable([]string{"pwsh"}, paths); pwshPath != "" {
		return chezmoi.Interpreter{
			Command: "pwsh",
			Args:    []string{"-NoLogo", "-File"},
		}
	}

	return chezmoi.Interpreter{}
}

func windowsVersion() (map[string]any, error) {
	return nil, nil
}
